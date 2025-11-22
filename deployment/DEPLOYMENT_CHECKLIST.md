# 🚀 生产环境部署检查清单

## 📋 部署前检查

### 1. 代码准备
- [ ] 所有代码已提交到 Git
- [ ] 通过所有测试
- [ ] 代码已经过 Code Review
- [ ] 版本号已更新

### 2. 环境变量配置
- [ ] `APP_ENV=production` 已设置
- [ ] `DATABASE_URL` 已正确配置
- [ ] `API_SERVER_PORT` 已设置（默认8080）
- [ ] 敏感信息已从代码中移除
- [ ] 所有必需的环境变量已配置

### 3. 数据库准备
- [ ] 生产数据库已创建
- [ ] 数据库表已初始化（运行 `init_database.sql`）
- [ ] 数据库连接已测试
- [ ] 数据库备份已配置
- [ ] 数据库用户权限已正确设置

### 4. 安全检查
- [ ] 使用强密码
- [ ] 启用 SSL/TLS 连接
- [ ] API 密钥已轮换
- [ ] 防火墙规则已配置
- [ ] 只开放必要的端口

### 5. 性能优化
- [ ] 编译时使用优化标志
- [ ] 数据库索引已创建
- [ ] 缓存策略已配置
- [ ] 日志级别设置为 `info` 或 `warn`

## 🔧 部署步骤

### 方式1: systemd 服务（推荐）

```bash
# 1. 编译应用
go build -o nofx main.go

# 2. 复制到部署目录
sudo cp nofx /opt/nofx/
sudo cp deployment/nofx.service /etc/systemd/system/

# 3. 编辑服务文件，设置环境变量
sudo nano /etc/systemd/system/nofx.service

# 4. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable nofx
sudo systemctl start nofx

# 5. 检查状态
sudo systemctl status nofx
```

**检查项**:
- [ ] 服务文件已正确配置
- [ ] 环境变量已设置
- [ ] 服务已启动
- [ ] 日志无错误

### 方式2: Docker

```bash
# 1. 构建镜像
docker build -f docker/Dockerfile.backend -t nofx:latest .

# 2. 启动容器
docker-compose -f deployment/docker-compose.production.yml up -d

# 3. 检查状态
docker ps
docker logs nofx-production
```

**检查项**:
- [ ] 镜像构建成功
- [ ] 容器正在运行
- [ ] 环境变量已传递
- [ ] 端口映射正确
- [ ] 健康检查通过

### 方式3: 直接运行

```bash
# 1. 编译
go build -o nofx main.go

# 2. 设置环境变量
export APP_ENV=production
export DATABASE_URL='postgresql://...'

# 3. 后台运行
nohup ./nofx > nofx.log 2>&1 &

# 4. 检查进程
ps aux | grep nofx
```

**检查项**:
- [ ] 进程正在运行
- [ ] 日志文件正常
- [ ] 端口已监听

## ✅ 部署后验证

### 1. 服务健康检查
```bash
# 检查健康端点
curl http://localhost:8080/api/health

# 应该返回 200 OK
```

### 2. 环境验证
```bash
# 查看日志，确认环境
# 应该看到: 🚀 运行环境: 生产环境 (production)
```

### 3. 数据库连接
```bash
# 日志中应该显示
# ✅ 数据库连接成功
# 🗄️ 数据库: postgresql://root:***@...
```

### 4. API 测试
```bash
# 测试基本 API
curl http://localhost:8080/api/supported-models
curl http://localhost:8080/api/supported-exchanges
```

### 5. 功能测试
- [ ] 用户可以登录
- [ ] 可以创建交易员
- [ ] AI 决策正常工作
- [ ] 数据正确保存到数据库

## 📊 监控设置

### 1. 日志监控
```bash
# systemd
sudo journalctl -u nofx -f

# Docker
docker logs -f nofx-production

# 文件
tail -f nofx.log
```

### 2. 性能监控
- [ ] CPU 使用率
- [ ] 内存使用率
- [ ] 磁盘使用率
- [ ] 网络流量

### 3. 应用监控
- [ ] API 响应时间
- [ ] 错误率
- [ ] 数据库连接数
- [ ] AI 请求成功率

## 🔄 回滚计划

如果部署出现问题：

### 快速回滚
```bash
# systemd
sudo systemctl stop nofx
sudo cp /opt/nofx/nofx.backup /opt/nofx/nofx
sudo systemctl start nofx

# Docker
docker-compose -f deployment/docker-compose.production.yml down
docker run -d --name nofx-production nofx:previous-version

# PM2
pm2 stop nofx-production
pm2 delete nofx-production
pm2 start nofx.backup
```

### 回滚检查
- [ ] 旧版本已恢复
- [ ] 服务正常运行
- [ ] 数据库兼容
- [ ] 用户功能正常

## 🛡️ 安全加固

### 1. 系统层面
- [ ] 禁用 root 登录
- [ ] 配置防火墙
- [ ] 启用 fail2ban
- [ ] 定期更新系统

### 2. 应用层面
- [ ] 使用非 root 用户运行
- [ ] 限制文件权限
- [ ] 启用 HTTPS
- [ ] 配置 CORS

### 3. 数据库层面
- [ ] 使用专用数据库用户
- [ ] 限制数据库访问 IP
- [ ] 启用 SSL 连接
- [ ] 定期备份

## 📝 文档更新

- [ ] 更新部署文档
- [ ] 记录配置变更
- [ ] 更新 API 文档
- [ ] 记录已知问题

## 🆘 故障处理

### 常见问题

#### 1. 服务无法启动
```bash
# 检查日志
sudo journalctl -u nofx -n 50

# 检查端口占用
lsof -i:8080

# 检查环境变量
sudo systemctl show nofx | grep Environment
```

#### 2. 数据库连接失败
```bash
# 测试连接
psql $DATABASE_URL -c "SELECT 1;"

# 检查防火墙
sudo ufw status

# 检查 PostgreSQL 服务
sudo systemctl status postgresql
```

#### 3. 内存不足
```bash
# 查看内存使用
free -h

# 重启服务
sudo systemctl restart nofx

# 或增加 swap
sudo fallocate -l 2G /swapfile
```

## 📞 联系方式

如果遇到问题：
1. 查看日志文件
2. 检查环境变量
3. 参考故障排除文档
4. 联系技术支持

## ✅ 最终检查

部署完成后，确认：
- [ ] 🚀 应用显示 "生产环境" 标识
- [ ] 🗄️ 连接到正确的数据库
- [ ] 📊 所有 API 正常响应
- [ ] 📝 日志正常记录
- [ ] 🔒 安全措施已启用
- [ ] 📈 监控已配置
- [ ] 💾 备份已设置
- [ ] 📚 文档已更新

---

**部署日期**: ___________
**部署人员**: ___________
**版本号**: ___________
**备注**: ___________
