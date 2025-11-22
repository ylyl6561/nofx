package config

import (
	"fmt"
)

// createTables 创建PostgreSQL数据库表
func (d *Database) createTables() error {
	// PostgreSQL 表结构
	queries := []string{
			// 系统级 AI 模型模板表（所有用户共享）
			`CREATE TABLE IF NOT EXISTS system_ai_models (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				provider TEXT NOT NULL,
				description TEXT DEFAULT '',
				default_model_name TEXT DEFAULT '',
				default_api_url TEXT DEFAULT '',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 系统级交易所模板表（所有用户共享）
			`CREATE TABLE IF NOT EXISTS system_exchanges (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				description TEXT DEFAULT '',
				supports_testnet BOOLEAN DEFAULT false,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// AI模型配置表
			`CREATE TABLE IF NOT EXISTS ai_models (
				id TEXT NOT NULL,
				user_id TEXT NOT NULL DEFAULT 'default',
				name TEXT NOT NULL,
				provider TEXT NOT NULL,
				enabled BOOLEAN DEFAULT false,
				api_key TEXT DEFAULT '',
				custom_api_url TEXT DEFAULT '',
				custom_model_name TEXT DEFAULT '',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id, user_id)
			)`,
			
			// 交易所配置表
			`CREATE TABLE IF NOT EXISTS exchanges (
				id TEXT NOT NULL,
				user_id TEXT NOT NULL DEFAULT 'default',
				api_key_name TEXT NOT NULL,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				enabled BOOLEAN DEFAULT false,
				api_key TEXT DEFAULT '',
				secret_key TEXT DEFAULT '',
				passphrase TEXT DEFAULT '',
				testnet BOOLEAN DEFAULT false,
				hyperliquid_wallet_addr TEXT DEFAULT '',
				aster_user TEXT DEFAULT '',
				aster_signer TEXT DEFAULT '',
				aster_private_key TEXT DEFAULT '',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id, user_id, api_key_name)
			)`,
			
			// 用户信号源配置表
			`CREATE TABLE IF NOT EXISTS user_signal_sources (
				id SERIAL PRIMARY KEY,
				user_id TEXT NOT NULL,
				coin_pool_url TEXT DEFAULT '',
				oi_top_url TEXT DEFAULT '',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(user_id)
			)`,
			
			// 交易员配置表
			`CREATE TABLE IF NOT EXISTS traders (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL DEFAULT 'default',
				name TEXT NOT NULL,
				ai_model_id TEXT NOT NULL,
				exchange_id TEXT NOT NULL,
				exchange_api_key_name TEXT DEFAULT '',
				initial_balance REAL NOT NULL,
				scan_interval_minutes INTEGER DEFAULT 3,
				is_running BOOLEAN DEFAULT false,
				btc_eth_leverage INTEGER DEFAULT 5,
				altcoin_leverage INTEGER DEFAULT 5,
				trading_symbols TEXT DEFAULT '',
				use_coin_pool BOOLEAN DEFAULT false,
				use_oi_top BOOLEAN DEFAULT false,
				custom_prompt TEXT DEFAULT '',
				override_base_prompt BOOLEAN DEFAULT false,
				is_cross_margin BOOLEAN DEFAULT true,
				use_default_coins BOOLEAN DEFAULT true,
				custom_coins TEXT DEFAULT '',
				system_prompt_template TEXT DEFAULT 'default',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 用户表
			`CREATE TABLE IF NOT EXISTS users (
				id TEXT PRIMARY KEY,
				email TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				otp_secret TEXT,
				otp_verified BOOLEAN DEFAULT false,
				is_admin BOOLEAN DEFAULT false,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 系统配置表
			`CREATE TABLE IF NOT EXISTS system_config (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 内测码表
			`CREATE TABLE IF NOT EXISTS beta_codes (
				code TEXT PRIMARY KEY,
				used BOOLEAN DEFAULT false,
				used_by TEXT DEFAULT '',
				used_at TIMESTAMP DEFAULT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// API Keys 表
			`CREATE TABLE IF NOT EXISTS api_keys (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				key_hash TEXT NOT NULL UNIQUE,
				key_prefix TEXT NOT NULL,
				name TEXT DEFAULT 'Default Key',
				enabled BOOLEAN DEFAULT true,
				rate_limit INTEGER DEFAULT 1000,
				usage_count INTEGER DEFAULT 0,
				last_used_at TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				expires_at TIMESTAMP
			)`,
			
			// API 使用记录表
			`CREATE TABLE IF NOT EXISTS api_usage_logs (
				id SERIAL PRIMARY KEY,
				api_key_id TEXT NOT NULL,
				user_id TEXT NOT NULL,
				endpoint TEXT NOT NULL,
				method TEXT NOT NULL,
				status_code INTEGER,
				response_time_ms INTEGER,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 用户交易配置表
			`CREATE TABLE IF NOT EXISTS user_trading_config (
				user_id TEXT PRIMARY KEY,
				btc_eth_leverage INTEGER DEFAULT NULL,
				altcoin_leverage INTEGER DEFAULT NULL,
				max_daily_loss REAL DEFAULT NULL,
				max_drawdown REAL DEFAULT NULL,
				stop_trading_minutes INTEGER DEFAULT NULL,
				use_default_coins BOOLEAN DEFAULT true,
				custom_coins TEXT DEFAULT '',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 决策日志表
			`CREATE TABLE IF NOT EXISTS decision_logs (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				trader_id TEXT NOT NULL,
				cycle_number INTEGER NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				system_prompt TEXT DEFAULT '',
				input_prompt TEXT DEFAULT '',
				cot_trace TEXT DEFAULT '',
				account_state TEXT NOT NULL,
				positions TEXT DEFAULT '',
				decisions TEXT DEFAULT '',
				candidate_coins TEXT DEFAULT '',
				execution_log TEXT DEFAULT '',
				success BOOLEAN DEFAULT false,
				error_message TEXT DEFAULT '',
				ai_request_duration_ms INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 权益历史表
			`CREATE TABLE IF NOT EXISTS equity_history (
				id SERIAL PRIMARY KEY,
				user_id TEXT NOT NULL,
				trader_id TEXT NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				total_equity REAL NOT NULL,
				available_balance REAL NOT NULL,
				total_unrealized_profit REAL NOT NULL,
				total_pnl REAL NOT NULL,
				total_pnl_pct REAL NOT NULL,
				position_count INTEGER DEFAULT 0,
				margin_used_pct REAL DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 决策日志索引
			`CREATE INDEX IF NOT EXISTS idx_decision_logs_user_trader 
				ON decision_logs(user_id, trader_id, timestamp DESC)`,
			`CREATE INDEX IF NOT EXISTS idx_decision_logs_timestamp 
				ON decision_logs(timestamp DESC)`,
			
			// 权益历史索引
			`CREATE INDEX IF NOT EXISTS idx_equity_history_user_trader 
				ON equity_history(user_id, trader_id, timestamp DESC)`,
			`CREATE INDEX IF NOT EXISTS idx_equity_history_timestamp 
				ON equity_history(timestamp DESC)`,
	}
	
	// 执行所有表创建语句
	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("创建表失败: %w\nSQL: %s", err, query)
		}
	}
	
	// 执行数据库迁移
	if err := d.runMigrations(); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}
	
	return nil
}

// runMigrations 执行PostgreSQL数据库迁移
func (d *Database) runMigrations() error {
	// PostgreSQL 迁移：添加 exchange_api_key_name 列到 traders 表
	migrations := []string{
		// 检查列是否存在，如果不存在则添加
		`DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name='traders' AND column_name='exchange_api_key_name'
			) THEN
				ALTER TABLE traders ADD COLUMN exchange_api_key_name TEXT DEFAULT '';
			END IF;
		END $$;`,
		// 添加 system_prompt, input_prompt, cot_trace 列到 decision_logs 表
		`DO $$ 
		BEGIN 
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name='decision_logs' AND column_name='system_prompt'
			) THEN
				ALTER TABLE decision_logs ADD COLUMN system_prompt TEXT DEFAULT '';
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name='decision_logs' AND column_name='input_prompt'
			) THEN
				ALTER TABLE decision_logs ADD COLUMN input_prompt TEXT DEFAULT '';
			END IF;
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name='decision_logs' AND column_name='cot_trace'
			) THEN
				ALTER TABLE decision_logs ADD COLUMN cot_trace TEXT DEFAULT '';
			END IF;
		END $$;`,
	}
	
	for _, migration := range migrations {
		if _, err := d.db.Exec(migration); err != nil {
			return fmt.Errorf("执行迁移失败: %w", err)
		}
	}
	
	return nil
}

