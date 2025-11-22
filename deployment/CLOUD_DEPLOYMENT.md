# 云平台部署指南

## 🌐 各云平台环境变量配置

### 1. Zeabur

#### 方式1: 通过Web界面
1. 登录 Zeabur Dashboard
2. 选择你的项目
3. 进入 "Variables" 标签
4. 添加环境变量：
   ```
   APP_ENV=production
   DATABASE_URL=postgresql://root:password@sjc1.clusters.zeabur.com:29853/zeabur?sslmode=disable
   API_SERVER_PORT=8080
   ```

#### 方式2: 使用 zeabur.json
```json
{
  "env": {
    "APP_ENV": "production",
    "DATABASE_URL": "${DATABASE_URL}",
    "API_SERVER_PORT": "8080"
  }
}
```

### 2. Vercel

在项目设置中添加环境变量：

```bash
# 通过CLI
vercel env add APP_ENV production
vercel env add DATABASE_URL production
vercel env add API_SERVER_PORT production

# 或在 vercel.json 中
{
  "env": {
    "APP_ENV": "production",
    "DATABASE_URL": "@database-url",
    "API_SERVER_PORT": "8080"
  }
}
```

### 3. Railway

#### 通过Web界面
1. 打开项目
2. 进入 "Variables" 标签
3. 添加变量

#### 通过CLI
```bash
railway variables set APP_ENV=production
railway variables set DATABASE_URL="postgresql://..."
railway variables set API_SERVER_PORT=8080
```

### 4. Heroku

```bash
# 通过CLI设置
heroku config:set APP_ENV=production
heroku config:set DATABASE_URL="postgresql://..."
heroku config:set API_SERVER_PORT=8080

# 查看配置
heroku config

# 或在 app.json 中
{
  "env": {
    "APP_ENV": {
      "value": "production"
    },
    "DATABASE_URL": {
      "required": true
    }
  }
}
```

### 5. AWS (EC2/ECS)

#### EC2 实例
在 `~/.bashrc` 或使用 systemd service 文件

#### ECS Task Definition
```json
{
  "containerDefinitions": [
    {
      "name": "nofx",
      "environment": [
        {
          "name": "APP_ENV",
          "value": "production"
        },
        {
          "name": "DATABASE_URL",
          "value": "postgresql://..."
        }
      ]
    }
  ]
}
```

#### 或使用 AWS Secrets Manager
```json
{
  "secrets": [
    {
      "name": "DATABASE_URL",
      "valueFrom": "arn:aws:secretsmanager:region:account:secret:nofx/database-url"
    }
  ]
}
```

### 6. Google Cloud Platform

#### Cloud Run
```bash
# 部署时设置
gcloud run deploy nofx \
  --set-env-vars APP_ENV=production \
  --set-env-vars DATABASE_URL="postgresql://..." \
  --set-env-vars API_SERVER_PORT=8080

# 或使用 env.yaml
env_variables:
  APP_ENV: production
  DATABASE_URL: postgresql://...
  API_SERVER_PORT: 8080
```

#### Compute Engine
在启动脚本中设置或使用 metadata

### 7. DigitalOcean App Platform

在 `.do/app.yaml` 中：
```yaml
name: nofx
services:
  - name: api
    envs:
      - key: APP_ENV
        value: production
      - key: DATABASE_URL
        value: ${DATABASE_URL}
      - key: API_SERVER_PORT
        value: "8080"
```

### 8. Azure

#### App Service
```bash
# 通过CLI
az webapp config appsettings set \
  --name nofx \
  --resource-group myResourceGroup \
  --settings APP_ENV=production DATABASE_URL="postgresql://..."

# 或在 Azure Portal 的 Configuration 中设置
```

## 🔐 敏感信息管理

### 使用密钥管理服务

#### AWS Secrets Manager
```go
import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/secretsmanager"
)

func getDatabaseURL() string {
    sess := session.Must(session.NewSession())
    svc := secretsmanager.New(sess)
    
    result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
        SecretId: aws.String("nofx/database-url"),
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    return *result.SecretString
}
```

#### HashiCorp Vault
```bash
# 存储密钥
vault kv put secret/nofx \
  database_url="postgresql://..." \
  api_key="sk-..."

# 读取密钥
vault kv get -field=database_url secret/nofx
```

## 📝 环境变量优先级

从高到低：

1. **命令行参数** (如果支持)
2. **系统环境变量** (`export`)
3. **云平台环境变量**
4. **`.env` 文件**
5. **代码中的默认值**

## 🛡️ 安全最佳实践

### 1. 不要在代码中硬编码
❌ 错误：
```go
dbURL := "postgresql://root:password@host/db"
```

✅ 正确：
```go
dbURL := os.Getenv("DATABASE_URL")
```

### 2. 使用不同的密钥
- 开发环境：弱密钥或测试密钥
- 生产环境：强密钥
- 绝不共享密钥

### 3. 定期轮换密钥
```bash
# 更新数据库密码后
heroku config:set DATABASE_URL="new_url"
# 或
kubectl set env deployment/nofx DATABASE_URL="new_url"
```

### 4. 限制访问权限
- 只有必要的人员能访问生产环境变量
- 使用 IAM 角色和权限管理
- 启用审计日志

### 5. 加密传输
- 使用 SSL/TLS 连接数据库
- HTTPS API 端点
- 加密存储的密钥

## 🔍 验证环境变量

部署后验证：

```bash
# 检查环境变量是否设置
curl http://your-app/api/health

# 查看日志确认环境
# 应该看到: 🚀 运行环境: 生产环境 (production)

# 检查数据库连接
# 应该看到: 🗄️ 数据库: postgresql://root:***@...
```

## 📚 相关资源

- [systemd 服务配置](../deployment/nofx.service)
- [Docker Compose 配置](../deployment/docker-compose.production.yml)
- [PM2 配置](../deployment/ecosystem.config.js)
- [环境变量说明](../docs/ENVIRONMENT_VARIABLES.md)
