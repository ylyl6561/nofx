-- 数据库迁移脚本：添加 token 统计和软删除功能
-- 执行方式: psql -d nofx_db -f scripts/migrate_add_tokens_and_soft_delete.sql

-- 1. 为 traders 表添加 total_tokens 字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'traders' AND column_name = 'total_tokens'
    ) THEN
        ALTER TABLE traders ADD COLUMN total_tokens BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_tokens 字段';
    ELSE
        RAISE NOTICE '⚠️  total_tokens 字段已存在，跳过';
    END IF;
END $$;

-- 2. 为 traders 表添加 is_deleted 字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'traders' AND column_name = 'is_deleted'
    ) THEN
        ALTER TABLE traders ADD COLUMN is_deleted TEXT DEFAULT 'n';
        RAISE NOTICE '✅ 已添加 is_deleted 字段 (TEXT DEFAULT ''n'')';
    ELSE
        RAISE NOTICE '⚠️  is_deleted 字段已存在，跳过';
    END IF;
END $$;

-- 3. 为 users 表添加 total_tokens 字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'total_tokens'
    ) THEN
        ALTER TABLE users ADD COLUMN total_tokens BIGINT DEFAULT 0;
        RAISE NOTICE '✅ 已添加 users.total_tokens 字段';
    ELSE
        RAISE NOTICE '⚠️  users.total_tokens 字段已存在，跳过';
    END IF;
END $$;

-- 4. 为 decision_logs 表添加 token 使用量字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'decision_logs' AND column_name = 'prompt_tokens'
    ) THEN
        ALTER TABLE decision_logs ADD COLUMN prompt_tokens INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 prompt_tokens 字段';
    ELSE
        RAISE NOTICE '⚠️  prompt_tokens 字段已存在，跳过';
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'decision_logs' AND column_name = 'completion_tokens'
    ) THEN
        ALTER TABLE decision_logs ADD COLUMN completion_tokens INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 completion_tokens 字段';
    ELSE
        RAISE NOTICE '⚠️  completion_tokens 字段已存在，跳过';
    END IF;
END $$;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'decision_logs' AND column_name = 'total_tokens'
    ) THEN
        ALTER TABLE decision_logs ADD COLUMN total_tokens INTEGER DEFAULT 0;
        RAISE NOTICE '✅ 已添加 total_tokens 字段';
    ELSE
        RAISE NOTICE '⚠️  total_tokens 字段已存在，跳过';
    END IF;
END $$;

-- 4. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_traders_is_deleted ON traders(is_deleted);
CREATE INDEX IF NOT EXISTS idx_traders_user_id_is_deleted ON traders(user_id, is_deleted);
CREATE INDEX IF NOT EXISTS idx_decision_logs_trader_id_timestamp ON decision_logs(trader_id, timestamp);

-- 5. 显示迁移完成信息
DO $$ 
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '✅ 数据库迁移完成！';
    RAISE NOTICE '========================================';
    RAISE NOTICE '新增字段:';
    RAISE NOTICE '  - traders.total_tokens (BIGINT)';
    RAISE NOTICE '  - traders.is_deleted (TEXT, ''n''/''y'')';
    RAISE NOTICE '  - users.total_tokens (BIGINT)';
    RAISE NOTICE '  - decision_logs.prompt_tokens (INTEGER)';
    RAISE NOTICE '  - decision_logs.completion_tokens (INTEGER)';
    RAISE NOTICE '  - decision_logs.total_tokens (INTEGER)';
    RAISE NOTICE '';
    RAISE NOTICE '新增索引:';
    RAISE NOTICE '  - idx_traders_is_deleted';
    RAISE NOTICE '  - idx_traders_user_id_is_deleted';
    RAISE NOTICE '  - idx_decision_logs_trader_id_timestamp';
    RAISE NOTICE '';
    RAISE NOTICE '⚠️  注意事项:';
    RAISE NOTICE '  - users.total_tokens 需要在代码中累加更新';
    RAISE NOTICE '  - 建议在 logger/db_decision_logger.go 中同时更新用户和trader的token';
    RAISE NOTICE '========================================';
END $$;
