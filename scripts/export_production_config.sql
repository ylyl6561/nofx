-- ============================================================
-- 导出生产环境配置数据
-- ============================================================
-- 使用方法: psql $PROD_DATABASE_URL -f export_production_config.sql
-- ============================================================

\echo '======================================'
\echo '📦 导出生产环境配置数据'
\echo '======================================'
\echo ''

-- 设置输出格式
\pset format unaligned
\pset tuples_only on
\pset fieldsep ','

-- 1. 导出系统AI模型模板
\echo '📤 导出 system_ai_models...'
\o system_ai_models.sql
SELECT 'INSERT INTO system_ai_models (id, name, provider, description, default_model_name, default_api_url, created_at) VALUES ('
    || quote_literal(id) || ', '
    || quote_literal(name) || ', '
    || quote_literal(provider) || ', '
    || quote_literal(description) || ', '
    || quote_literal(default_model_name) || ', '
    || quote_literal(default_api_url) || ', '
    || quote_literal(created_at::text) || ') ON CONFLICT (id) DO UPDATE SET '
    || 'name = EXCLUDED.name, '
    || 'provider = EXCLUDED.provider, '
    || 'description = EXCLUDED.description, '
    || 'default_model_name = EXCLUDED.default_model_name, '
    || 'default_api_url = EXCLUDED.default_api_url;'
FROM system_ai_models;
\o

-- 2. 导出系统交易所模板
\echo '📤 导出 system_exchanges...'
\o system_exchanges.sql
SELECT 'INSERT INTO system_exchanges (id, name, type, description, default_api_url, default_testnet_url, created_at) VALUES ('
    || quote_literal(id) || ', '
    || quote_literal(name) || ', '
    || quote_literal(type) || ', '
    || quote_literal(description) || ', '
    || quote_literal(default_api_url) || ', '
    || quote_literal(default_testnet_url) || ', '
    || quote_literal(created_at::text) || ') ON CONFLICT (id) DO UPDATE SET '
    || 'name = EXCLUDED.name, '
    || 'type = EXCLUDED.type, '
    || 'description = EXCLUDED.description, '
    || 'default_api_url = EXCLUDED.default_api_url, '
    || 'default_testnet_url = EXCLUDED.default_testnet_url;'
FROM system_exchanges;
\o

-- 3. 导出系统配置
\echo '📤 导出 system_config...'
\o system_config.sql
SELECT 'INSERT INTO system_config (key, value, updated_at) VALUES ('
    || quote_literal(key) || ', '
    || quote_literal(value) || ', '
    || quote_literal(updated_at::text) || ') ON CONFLICT (key) DO UPDATE SET '
    || 'value = EXCLUDED.value, '
    || 'updated_at = EXCLUDED.updated_at;'
FROM system_config;
\o

-- 4. 导出内测码（可选）
\echo '📤 导出 beta_codes...'
\o beta_codes.sql
SELECT 'INSERT INTO beta_codes (code, used, used_by, used_at, created_at) VALUES ('
    || quote_literal(code) || ', '
    || used || ', '
    || COALESCE(quote_literal(used_by), 'NULL') || ', '
    || COALESCE(quote_literal(used_at::text), 'NULL') || ', '
    || quote_literal(created_at::text) || ') ON CONFLICT (code) DO NOTHING;'
FROM beta_codes
WHERE NOT used;  -- 只导出未使用的内测码
\o

\echo ''
\echo '======================================'
\echo '✅ 导出完成！'
\echo '======================================'
\echo ''
\echo '生成的文件:'
\echo '  - system_ai_models.sql'
\echo '  - system_exchanges.sql'
\echo '  - system_config.sql'
\echo '  - beta_codes.sql'
\echo ''
\echo '导入到本地数据库:'
\echo '  psql $LOCAL_DATABASE_URL -f system_ai_models.sql'
\echo '  psql $LOCAL_DATABASE_URL -f system_exchanges.sql'
\echo '  psql $LOCAL_DATABASE_URL -f system_config.sql'
\echo '  psql $LOCAL_DATABASE_URL -f beta_codes.sql'
\echo ''
