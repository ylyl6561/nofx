-- ============================================================
-- NoFX 常用SQL查询
-- ============================================================
-- 使用方法: 
--   1. 连接数据库: psql $DATABASE_URL
--   2. 执行查询: \i scripts/useful_queries.sql
--   3. 或复制粘贴单个查询
-- ============================================================

-- ============================================================
-- 1. 系统概览
-- ============================================================

-- 查看所有表及记录数
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    n_live_tup AS row_count
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 查看数据库大小
SELECT 
    pg_database.datname,
    pg_size_pretty(pg_database_size(pg_database.datname)) AS size
FROM pg_database
WHERE datname = current_database();

-- ============================================================
-- 2. 用户相关
-- ============================================================

-- 查看所有用户
SELECT 
    id,
    email,
    is_admin,
    otp_verified,
    created_at
FROM users
ORDER BY created_at DESC;

-- 统计每个用户的交易员数量
SELECT 
    u.email,
    COUNT(t.id) as trader_count
FROM users u
LEFT JOIN traders t ON u.id = t.user_id
GROUP BY u.id, u.email
ORDER BY trader_count DESC;

-- ============================================================
-- 3. 交易员相关
-- ============================================================

-- 查看所有交易员
SELECT 
    id,
    user_id,
    name,
    ai_model_id,
    exchange_id,
    initial_balance,
    is_running,
    created_at
FROM traders
ORDER BY created_at DESC;

-- 查看正在运行的交易员
SELECT 
    t.id,
    t.name,
    u.email as user_email,
    t.ai_model_id,
    t.exchange_id,
    t.initial_balance
FROM traders t
JOIN users u ON t.user_id = u.id
WHERE t.is_running = true;

-- ============================================================
-- 4. 决策日志分析
-- ============================================================

-- 最近10条决策日志
SELECT 
    trader_id,
    cycle_number,
    timestamp,
    success,
    error_message,
    ai_request_duration_ms
FROM decision_logs
ORDER BY timestamp DESC
LIMIT 10;

-- 每个交易员的决策统计
SELECT 
    trader_id,
    COUNT(*) as total_decisions,
    SUM(CASE WHEN success THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN NOT success THEN 1 ELSE 0 END) as failed,
    ROUND(100.0 * SUM(CASE WHEN success THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate,
    ROUND(AVG(ai_request_duration_ms), 2) as avg_ai_duration_ms
FROM decision_logs
GROUP BY trader_id
ORDER BY total_decisions DESC;

-- 查看包含完整prompt和AI响应的决策
SELECT 
    trader_id,
    timestamp,
    LENGTH(system_prompt) as system_prompt_length,
    LENGTH(input_prompt) as input_prompt_length,
    LENGTH(cot_trace) as cot_trace_length,
    success
FROM decision_logs
WHERE system_prompt != '' OR input_prompt != '' OR cot_trace != ''
ORDER BY timestamp DESC
LIMIT 10;

-- 今天的决策统计
SELECT 
    trader_id,
    COUNT(*) as decisions_today,
    SUM(CASE WHEN success THEN 1 ELSE 0 END) as successful,
    ROUND(AVG(ai_request_duration_ms), 2) as avg_duration_ms
FROM decision_logs
WHERE DATE(timestamp) = CURRENT_DATE
GROUP BY trader_id;

-- ============================================================
-- 5. 权益历史分析
-- ============================================================

-- 最近权益记录
SELECT 
    trader_id,
    timestamp,
    total_equity,
    total_pnl,
    total_pnl_pct,
    position_count
FROM equity_history
ORDER BY timestamp DESC
LIMIT 20;

-- 每个交易员的收益统计
SELECT 
    trader_id,
    COUNT(*) as record_count,
    MIN(total_equity) as min_equity,
    MAX(total_equity) as max_equity,
    ROUND(AVG(total_equity), 2) as avg_equity,
    ROUND(MIN(total_pnl_pct), 2) as min_pnl_pct,
    ROUND(MAX(total_pnl_pct), 2) as max_pnl_pct
FROM equity_history
GROUP BY trader_id
ORDER BY max_pnl_pct DESC;

-- 今天的权益变化
SELECT 
    trader_id,
    MIN(total_equity) as start_equity,
    MAX(total_equity) as current_equity,
    ROUND(MAX(total_equity) - MIN(total_equity), 2) as change,
    ROUND(100.0 * (MAX(total_equity) - MIN(total_equity)) / MIN(total_equity), 2) as change_pct
FROM equity_history
WHERE DATE(timestamp) = CURRENT_DATE
GROUP BY trader_id;

-- ============================================================
-- 6. AI模型和交易所配置
-- ============================================================

-- 查看AI模型配置
SELECT 
    id,
    user_id,
    name,
    provider,
    enabled,
    custom_model_name,
    created_at
FROM ai_models
ORDER BY created_at DESC;

-- 查看交易所配置
SELECT 
    id,
    user_id,
    name,
    type,
    enabled,
    testnet,
    created_at
FROM exchanges
ORDER BY created_at DESC;

-- 统计每种AI模型的使用情况
SELECT 
    am.provider,
    COUNT(DISTINCT t.id) as trader_count,
    COUNT(DISTINCT t.user_id) as user_count
FROM ai_models am
LEFT JOIN traders t ON am.id = t.ai_model_id
WHERE am.enabled = true
GROUP BY am.provider;

-- ============================================================
-- 7. 性能分析
-- ============================================================

-- 最慢的AI请求
SELECT 
    trader_id,
    timestamp,
    ai_request_duration_ms,
    success,
    error_message
FROM decision_logs
WHERE ai_request_duration_ms > 0
ORDER BY ai_request_duration_ms DESC
LIMIT 10;

-- AI请求时长分布
SELECT 
    CASE 
        WHEN ai_request_duration_ms < 1000 THEN '< 1s'
        WHEN ai_request_duration_ms < 5000 THEN '1-5s'
        WHEN ai_request_duration_ms < 10000 THEN '5-10s'
        WHEN ai_request_duration_ms < 30000 THEN '10-30s'
        ELSE '> 30s'
    END as duration_range,
    COUNT(*) as count
FROM decision_logs
WHERE ai_request_duration_ms > 0
GROUP BY duration_range
ORDER BY 
    CASE duration_range
        WHEN '< 1s' THEN 1
        WHEN '1-5s' THEN 2
        WHEN '5-10s' THEN 3
        WHEN '10-30s' THEN 4
        ELSE 5
    END;

-- ============================================================
-- 8. 数据清理（谨慎使用）
-- ============================================================

-- 查看旧数据（不删除）
SELECT 
    'decision_logs' as table_name,
    COUNT(*) as old_records
FROM decision_logs
WHERE created_at < NOW() - INTERVAL '30 days'
UNION ALL
SELECT 
    'equity_history' as table_name,
    COUNT(*) as old_records
FROM equity_history
WHERE created_at < NOW() - INTERVAL '90 days';

-- 删除旧决策日志（30天前）
-- DELETE FROM decision_logs WHERE created_at < NOW() - INTERVAL '30 days';

-- 删除旧权益历史（90天前）
-- DELETE FROM equity_history WHERE created_at < NOW() - INTERVAL '90 days';

-- ============================================================
-- 9. 实用工具查询
-- ============================================================

-- 查看当前连接信息
SELECT 
    current_database() as database,
    current_user as user,
    inet_server_addr() as server_ip,
    inet_server_port() as server_port,
    version() as pg_version;

-- 查看活动连接
SELECT 
    pid,
    usename,
    application_name,
    client_addr,
    state,
    query_start,
    state_change
FROM pg_stat_activity
WHERE datname = current_database()
ORDER BY query_start DESC;

-- 查看表的最后更新时间
SELECT 
    schemaname,
    tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    last_vacuum,
    last_autovacuum
FROM pg_stat_user_tables
ORDER BY n_tup_ins + n_tup_upd + n_tup_del DESC;

-- ============================================================
-- 提示：
-- 1. 使用 \x 切换扩展显示模式（适合宽表）
-- 2. 使用 \timing 显示查询执行时间
-- 3. 使用 LIMIT 限制结果数量
-- 4. 生产环境谨慎使用 DELETE 操作
-- ============================================================
