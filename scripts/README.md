# Scripts 目录说明

本目录包含用于数据库管理和应用启动的各种脚本。

## 📋 脚本列表

### 数据库相关

#### `init_database.sql`
完整的数据库初始化SQL脚本，包含所有14个表的创建语句。

**用途**: 初始化或重置数据库结构

**使用方法**:
```bash
psql $DATABASE_URL -f scripts/init_database.sql
```

#### `init_db.sh`
自动化的数据库初始化脚本，包含连接检查和错误处理。

**用途**: 一键初始化数据库

**使用方法**:
```bash
export DATABASE_URL='postgresql://user:pass@host:port/db'
chmod +x scripts/init_db.sh
./scripts/init_db.sh
```

#### `setup_local_db.sh`
快速设置本地测试数据库的脚本。

**用途**: 创建本地PostgreSQL测试数据库并初始化表结构

**使用方法**:
```bash
chmod +x scripts/setup_local_db.sh
./scripts/setup_local_db.sh
```

**功能**:
- 检查PostgreSQL服务状态
- 创建 `nofx_test` 数据库
- 初始化所有表和索引
- 生成连接信息

### 应用启动相关

#### `run_local.sh`
启动本地测试环境的脚本。

**用途**: 使用本地数据库启动应用

**使用方法**:
```bash
chmod +x scripts/run_local.sh
./scripts/run_local.sh
```

**特点**:
- 自动加载 `.env.local` 配置
- 检查数据库连接
- 初始化数据库表（如果需要）
- 启动应用

#### `run_production.sh`
启动生产环境的脚本。

**用途**: 使用生产数据库启动应用

**使用方法**:
```bash
chmod +x scripts/run_production.sh
./scripts/run_production.sh
```

**特点**:
- 自动加载 `.env.production` 配置
- 检查数据库连接
- 显示生产环境警告
- 启动应用

## 🚀 快速开始

### 首次设置本地环境

```bash
# 1. 设置本地数据库
./scripts/setup_local_db.sh

# 2. 编辑环境配置
nano .env.local
# 填入: DATABASE_URL=postgresql://postgres:password@localhost:5432/nofx_test

# 3. 启动应用
./scripts/run_local.sh
```

### 首次设置生产环境

```bash
# 1. 创建生产环境配置
nano .env.production
# 填入生产数据库连接信息

# 2. 初始化生产数据库
export DATABASE_URL='your_production_url'
./scripts/init_db.sh

# 3. 启动应用
./scripts/run_production.sh
```

## 📊 环境对比

| 特性 | 本地环境 | 生产环境 |
|------|---------|---------|
| 数据库 | localhost | 远程服务器 |
| 配置文件 | `.env.local` | `.env.production` |
| 启动脚本 | `run_local.sh` | `run_production.sh` |
| 数据隔离 | ✅ 完全隔离 | ⚠️ 生产数据 |
| 响应速度 | 🚀 快速 | 🐢 取决于网络 |

## 🔧 常用命令

### 数据库管理

```bash
# 连接本地数据库
psql postgresql://postgres@localhost:5432/nofx_test

# 查看所有表
psql $DATABASE_URL -c "\dt"

# 备份数据库
pg_dump $DATABASE_URL > backup.sql

# 恢复数据库
psql $DATABASE_URL < backup.sql

# 重置数据库
dropdb nofx_test && createdb nofx_test
psql -d nofx_test -f scripts/init_database.sql
```

### 环境切换

```bash
# 切换到本地环境
export $(cat .env.local | xargs)

# 切换到生产环境
export $(cat .env.production | xargs)

# 查看当前数据库
echo $DATABASE_URL
```

## 🛠️ 故障排除

### 脚本权限问题

```bash
# 给所有脚本添加执行权限
chmod +x scripts/*.sh
```

### PostgreSQL未启动

```bash
# macOS
brew services start postgresql@14

# Linux
sudo systemctl start postgresql
```

### 数据库连接失败

```bash
# 检查PostgreSQL服务
pg_isready

# 测试连接
psql $DATABASE_URL -c "SELECT 1;"
```

### 端口被占用

```bash
# 查看占用端口的进程
lsof -ti:8080

# 杀死进程
lsof -ti:8080 | xargs kill -9
```

## 📝 最佳实践

1. **开发时使用本地环境**
   ```bash
   ./scripts/run_local.sh
   ```

2. **测试前备份数据**
   ```bash
   pg_dump nofx_test > backup_before_test.sql
   ```

3. **定期更新数据库结构**
   ```bash
   psql $DATABASE_URL -f scripts/init_database.sql
   ```

4. **不要提交环境配置文件**
   - `.env.local` 和 `.env.production` 已在 `.gitignore` 中

## 📚 相关文档

- [数据库设置指南](DATABASE_SETUP.md)
- [本地开发指南](../docs/LOCAL_DEVELOPMENT.md)
- [快速开始指南](../QUICK_START_GUIDE.md)

## 🆘 获取帮助

如果遇到问题：

1. 查看脚本输出的错误信息
2. 检查 `.env.local` 或 `.env.production` 配置
3. 验证PostgreSQL服务状态
4. 查阅相关文档
5. 提交Issue到项目仓库
