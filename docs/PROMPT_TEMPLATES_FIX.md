# Prompt Templates Deployment Fix

## 问题描述
部署后 `/api/prompt-templates` 接口无法正常工作，返回空列表或错误。

## 根本原因
`prompts` 目录在 Docker 镜像构建时没有被复制到最终镜像中，导致应用启动时无法加载提示词模板文件。

## 解决方案

### 1. 修改根目录 Dockerfile
在 `Dockerfile` 的最终阶段添加 prompts 目录的复制：

```dockerfile
FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /src/app /app/app
COPY --from=backend-builder /src/web/dist /app/web/dist
COPY --from=backend-builder /src/prompts /app/prompts  # ← 新增此行
EXPOSE 8080
CMD ["/app/app"]
```

### 2. 修改 docker/Dockerfile.backend
在 `docker/Dockerfile.backend` 的运行时阶段添加 prompts 目录：

```dockerfile
COPY --from=ta-lib-builder /usr/local /usr/local
WORKDIR /app
COPY --from=backend-builder /app/nofx .
COPY --from=backend-builder /app/prompts ./prompts  # ← 新增此行

EXPOSE 8080
```

### 3. Docker Compose 配置
`docker-compose.yml` 已经正确配置了 volume 挂载：
```yaml
volumes:
  - ./prompts:/app/prompts  # ✓ 已存在
```

### 4. Zeabur 部署
使用 `zbpack.json` 的部署方式，prompts 目录应该会自动包含在部署包中。如果仍有问题，确保：
- prompts 目录不在 `.gitignore` 中
- 所有 .txt 文件都已提交到 git 仓库

## 验证步骤

### 本地测试
```bash
# 重新构建 Docker 镜像
docker build -t nofx:test .

# 运行容器
docker run -p 8080:8080 nofx:test

# 测试接口
curl http://localhost:8080/api/prompt-templates
```

### 部署后测试
```bash
curl https://your-domain.com/api/prompt-templates
```

预期返回：
```json
{
  "templates": [
    {"name": "default"},
    {"name": "adaptive"},
    {"name": "Hansen"},
    ...
  ]
}
```

## 相关文件
- `/Users/yuliang/project/nofx/Dockerfile` - 已修复
- `/Users/yuliang/project/nofx/docker/Dockerfile.backend` - 已修复
- `/Users/yuliang/project/nofx/decision/prompt_manager.go` - 提示词加载逻辑
- `/Users/yuliang/project/nofx/api/server.go` - API 端点实现

## 注意事项
1. 修改后需要重新构建 Docker 镜像
2. 如果使用 Zeabur 部署，需要重新触发部署
3. prompts 目录中的所有 .txt 文件都会被加载为模板
4. 模板名称为文件名（不含 .txt 扩展名）
