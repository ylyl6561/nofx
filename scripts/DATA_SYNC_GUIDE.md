# 数据同步指南

## 📋 概述

本指南说明如何从生产环境同步配置数据到本地测试环境。

## 🎯 同步的数据

### 会同步的表
- ✅ `system_ai_models` - 系统AI模型模板
- ✅ `system_exchanges` - 系统交易所模板
- ✅ `system_config` - 系统配置
- ✅ `beta_codes` - 内测码（仅未使用的）

### 不会同步的表
- ❌ `users` - 用户数据（避免冲突）
- ❌ `traders` - 交易员配置（避免冲突）
- ❌ `ai_models` - 用户AI模型配置
- ❌ `exchanges` - 用户交易所配置
- ❌ `decision_logs` - 决策日志（数据量大）
- ❌ `equity_history` - 权益历史（数据量大）
- ❌ `api_keys` - API密钥（安全考虑）
- ❌ `api_usage_logs` - API使用日志

## 🚀 使用方法

### 方法1: 使用Shell脚本（推荐）

```bash
# 1. 给脚本添加执行权限
chmod +x scripts/sync_config_from_production.sh

# 2. 运行同步脚本
./scripts/sync_config_from_production.sh
```

**优点**: 简单快速，无需额外依赖

### 方法2: 使用Python脚本

```bash
# 1. 安装依赖
pip3 install psycopg2-binary

# 2. 给脚本添加执行权限
chmod +x scripts/sync_data.py

# 3. 运行同步脚本
./scripts/sync_data.py
```

**优点**: 更灵活，错误处理更好

### 方法3: 手动SQL导出导入

```bash
# 1. 导出生产环境数据
psql $PROD_DATABASE_URL -f scripts/export_production_config.sql

# 2. 导入到本地环境
psql $LOCAL_DATABASE_URL -f system_ai_models.sql
psql $LOCAL_DATABASE_URL -f system_exchanges.sql
psql $LOCAL_DATABASE_URL -f system_config.sql
psql $LOCAL_DATABASE_URL -f beta_codes.sql
```

**优点**: 完全可控，可以选择性导入

## 📝 详细步骤

### 准备工作

1. **确保本地PostgreSQL已启动**
   ```bash
   brew services start postgresql@14
   ```

2. **确保本地数据库已创建**
   ```bash
   createdb nofx_test
   psql -d nofx_test -f scripts/init_database.sql
   ```

3. **确保环境配置文件存在**
   - `.env.local` - 本地数据库配置
   - `.env.production` - 生产数据库配置

### 执行同步

#### 使用Shell脚本

```bash
./scripts/sync_config_from_production.sh
```

**输出示例**:
```
🔄 配置数据同步工具
======================================

📋 加载生产环境配置...
📋 加载本地环境配置...

📊 数据库连接信息:
  生产: postgresql://root:***@sjc1.clusters.zeabur.com...
  本地: postgresql://postgres:***@localhost:5432/nofx_test...

======================================
📦 开始导出生产环境配置数据...
======================================

📤 导出表: system_ai_models
   ✓ 导出 3 条记录
📤 导出表: system_exchanges
   ✓ 导出 5 条记录
📤 导出表: system_config
   ✓ 导出 12 条记录

======================================
📥 开始导入到本地测试环境...
======================================

⚠️  这将覆盖本地数据库中的配置数据，是否继续? (y/n) y

📥 导入表: system_ai_models
   ✓ 导入 3 条记录
📥 导入表: system_exchanges
   ✓ 导入 5 条记录
📥 导入表: system_config
   ✓ 导入 12 条记录

======================================
✅ 配置数据同步完成！
======================================
```

#### 使用Python脚本

```bash
./scripts/sync_data.py
```

**输出示例**:
```
🔄 配置数据同步工具 (Python版)
============================================================

📋 加载环境配置...
  生产: postgresql://root:***@sjc1.clusters.zeabur.com...
  本地: postgresql://postgres:***@localhost:5432/nofx_test...

🔌 连接数据库...
  ✓ 连接成功

⚠️  这将覆盖本地数据库中的配置数据
是否继续? (y/n): y

============================================================
📦 开始同步数据...
============================================================

📊 处理表: system_ai_models (系统AI模型模板)
   📤 从生产环境导出...
   ✓ 导出 3 条记录
   📥 导入到本地环境...
   ✓ 导入 3 条记录

============================================================
✅ 同步完成！
============================================================

📊 统计:
  导出: 20 条记录
  导入: 20 条记录
```

## 🔍 验证同步结果

### 方法1: 使用psql

```bash
# 连接本地数据库
psql postgresql://postgres@localhost:5432/nofx_test

# 查看系统AI模型
SELECT * FROM system_ai_models;

# 查看系统交易所
SELECT * FROM system_exchanges;

# 查看系统配置
SELECT * FROM system_config;

# 退出
\q
```

### 方法2: 使用数据库客户端

使用 TablePlus、DBeaver 等工具连接本地数据库，查看表数据。

### 方法3: 启动应用验证

```bash
# 启动本地环境
./scripts/run_local.sh

# 访问 API 查看配置
curl http://localhost:8080/api/supported-models
curl http://localhost:8080/api/supported-exchanges
```

## ⚠️ 注意事项

### 1. 数据覆盖警告
- 同步会**清空并覆盖**本地数据库中的配置表
- 如有重要的本地配置，请先备份

### 2. 用户数据隔离
- 用户数据（users, traders）**不会**同步
- 避免本地测试影响生产用户

### 3. 敏感信息
- API密钥**不会**同步
- 需要在本地单独配置

### 4. 数据量考虑
- 决策日志和权益历史**不会**同步
- 如需同步，请手动修改脚本

## 🔧 自定义同步

### 添加新表到同步列表

编辑 `sync_config_from_production.sh`:

```bash
TABLES=(
    "system_ai_models"
    "system_exchanges"
    "system_config"
    "your_new_table"  # 添加新表
)
```

或编辑 `sync_data.py`:

```python
SYNC_TABLES = {
    'your_new_table': {
        'columns': ['id', 'name', 'value'],
        'primary_key': 'id',
        'description': '你的新表'
    }
}
```

### 添加过滤条件

在 `sync_data.py` 中添加 `filter` 字段:

```python
'beta_codes': {
    'columns': ['code', 'used', 'used_by', 'used_at', 'created_at'],
    'primary_key': 'code',
    'description': '内测码',
    'filter': 'used = false AND created_at > NOW() - INTERVAL \'30 days\''
}
```

## 🛠️ 故障排除

### 问题1: 数据库连接失败

**错误**: `❌ 数据库连接失败`

**解决方案**:
```bash
# 检查生产数据库连接
psql $PROD_DATABASE_URL -c "SELECT 1;"

# 检查本地数据库连接
psql $LOCAL_DATABASE_URL -c "SELECT 1;"

# 检查环境配置文件
cat .env.production
cat .env.local
```

### 问题2: 表不存在

**错误**: `relation "table_name" does not exist`

**解决方案**:
```bash
# 重新初始化本地数据库
psql -d nofx_test -f scripts/init_database.sql
```

### 问题3: Python依赖缺失

**错误**: `ModuleNotFoundError: No module named 'psycopg2'`

**解决方案**:
```bash
pip3 install psycopg2-binary
```

### 问题4: 权限不足

**错误**: `permission denied`

**解决方案**:
```bash
# 给脚本添加执行权限
chmod +x scripts/sync_config_from_production.sh
chmod +x scripts/sync_data.py
```

## 📚 相关文档

- [本地开发指南](../docs/LOCAL_DEVELOPMENT.md)
- [数据库设置指南](DATABASE_SETUP.md)
- [环境变量配置](../docs/ENVIRONMENT_VARIABLES.md)

## 💡 最佳实践

1. **定期同步** - 生产环境配置更新后及时同步
2. **同步前备份** - 如有重要本地配置，先备份
3. **验证结果** - 同步后验证数据完整性
4. **选择性同步** - 只同步需要的表，避免不必要的数据传输
5. **使用脚本** - 使用自动化脚本，减少人为错误

## 🔄 定期同步建议

```bash
# 创建定时任务（可选）
# 每天凌晨2点同步配置
0 2 * * * cd /path/to/nofx && ./scripts/sync_config_from_production.sh > /tmp/sync.log 2>&1
```

---

**最后更新**: 2025-11-22
