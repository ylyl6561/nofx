package config

import (
	"fmt"
)

// createTables 创建数据库表
func (d *Database) createTables() error {
	var queries []string
	
	if d.isPostgreSQL() {
		// PostgreSQL 表结构
		queries = []string{
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
		}
	} else {
		// SQLite 表结构
		queries = []string{
			// 系统级 AI 模型模板表（所有用户共享）
			`CREATE TABLE IF NOT EXISTS system_ai_models (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				provider TEXT NOT NULL,
				description TEXT DEFAULT '',
				default_model_name TEXT DEFAULT '',
				default_api_url TEXT DEFAULT '',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 系统级交易所模板表（所有用户共享）
			`CREATE TABLE IF NOT EXISTS system_exchanges (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				description TEXT DEFAULT '',
				supports_testnet BOOLEAN DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// AI模型配置表
			`CREATE TABLE IF NOT EXISTS ai_models (
				id TEXT NOT NULL,
				user_id TEXT NOT NULL DEFAULT 'default',
				name TEXT NOT NULL,
				provider TEXT NOT NULL,
				enabled BOOLEAN DEFAULT 0,
				api_key TEXT DEFAULT '',
				custom_api_url TEXT DEFAULT '',
				custom_model_name TEXT DEFAULT '',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id, user_id)
			)`,
			
			// 交易所配置表
			`CREATE TABLE IF NOT EXISTS exchanges (
				id TEXT NOT NULL,
				user_id TEXT NOT NULL DEFAULT 'default',
				api_key_name TEXT NOT NULL,
				name TEXT NOT NULL,
				type TEXT NOT NULL,
				enabled BOOLEAN DEFAULT 0,
				api_key TEXT DEFAULT '',
				secret_key TEXT DEFAULT '',
				passphrase TEXT DEFAULT '',
				testnet BOOLEAN DEFAULT 0,
				hyperliquid_wallet_addr TEXT DEFAULT '',
				aster_user TEXT DEFAULT '',
				aster_signer TEXT DEFAULT '',
				aster_private_key TEXT DEFAULT '',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id, user_id, api_key_name)
			)`,
			
			// 用户信号源配置表
			`CREATE TABLE IF NOT EXISTS user_signal_sources (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id TEXT NOT NULL,
				coin_pool_url TEXT DEFAULT '',
				oi_top_url TEXT DEFAULT '',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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
				is_running BOOLEAN DEFAULT 0,
				btc_eth_leverage INTEGER DEFAULT 5,
				altcoin_leverage INTEGER DEFAULT 5,
				trading_symbols TEXT DEFAULT '',
				use_coin_pool BOOLEAN DEFAULT 0,
				use_oi_top BOOLEAN DEFAULT 0,
				custom_prompt TEXT DEFAULT '',
				override_base_prompt BOOLEAN DEFAULT 0,
				is_cross_margin BOOLEAN DEFAULT 1,
				use_default_coins BOOLEAN DEFAULT 1,
				custom_coins TEXT DEFAULT '',
				system_prompt_template TEXT DEFAULT 'default',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 用户表
			`CREATE TABLE IF NOT EXISTS users (
				id TEXT PRIMARY KEY,
				email TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				otp_secret TEXT,
				otp_verified BOOLEAN DEFAULT 0,
				is_admin BOOLEAN DEFAULT 0,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 系统配置表
			`CREATE TABLE IF NOT EXISTS system_config (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// 内测码表
			`CREATE TABLE IF NOT EXISTS beta_codes (
				code TEXT PRIMARY KEY,
				used BOOLEAN DEFAULT 0,
				used_by TEXT DEFAULT '',
				used_at DATETIME DEFAULT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// API Keys 表
			`CREATE TABLE IF NOT EXISTS api_keys (
				id TEXT PRIMARY KEY,
				user_id TEXT NOT NULL,
				key_hash TEXT NOT NULL UNIQUE,
				key_prefix TEXT NOT NULL,
				name TEXT DEFAULT 'Default Key',
				enabled BOOLEAN DEFAULT 1,
				rate_limit INTEGER DEFAULT 1000,
				usage_count INTEGER DEFAULT 0,
				last_used_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				expires_at DATETIME
			)`,
			
			// API 使用记录表
			`CREATE TABLE IF NOT EXISTS api_usage_logs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				api_key_id TEXT NOT NULL,
				user_id TEXT NOT NULL,
				endpoint TEXT NOT NULL,
				method TEXT NOT NULL,
				status_code INTEGER,
				response_time_ms INTEGER,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			
			// SQLite 触发器：自动更新 updated_at
			`CREATE TRIGGER IF NOT EXISTS update_users_updated_at
				AFTER UPDATE ON users
				BEGIN
					UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END`,
			
			`CREATE TRIGGER IF NOT EXISTS update_ai_models_updated_at
				AFTER UPDATE ON ai_models
				BEGIN
					UPDATE ai_models SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id AND user_id = NEW.user_id;
				END`,
			
			`CREATE TRIGGER IF NOT EXISTS update_exchanges_updated_at
				AFTER UPDATE ON exchanges
				BEGIN
					UPDATE exchanges SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id AND user_id = NEW.user_id;
				END`,
			
			`CREATE TRIGGER IF NOT EXISTS update_traders_updated_at
				AFTER UPDATE ON traders
				BEGIN
					UPDATE traders SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END`,
			
			`CREATE TRIGGER IF NOT EXISTS update_user_signal_sources_updated_at
				AFTER UPDATE ON user_signal_sources
				BEGIN
					UPDATE user_signal_sources SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
				END`,
			
			`CREATE TRIGGER IF NOT EXISTS update_system_config_updated_at
				AFTER UPDATE ON system_config
				BEGIN
					UPDATE system_config SET updated_at = CURRENT_TIMESTAMP WHERE key = NEW.key;
				END`,
		}
	}
	
	// 执行建表语句
	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("执行SQL失败 [%s]: %w", query, err)
		}
	}
	
	// 执行数据库迁移
	if err := d.runMigrations(); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}
	
	return nil
}

// runMigrations 执行数据库迁移
func (d *Database) runMigrations() error {
	if d.isPostgreSQL() {
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
		}
		
		for _, migration := range migrations {
			if _, err := d.db.Exec(migration); err != nil {
				return fmt.Errorf("执行迁移失败 [%s]: %w", migration, err)
			}
		}
	} else {
		// SQLite 迁移：添加 exchange_api_key_name 列到 traders 表
		// SQLite 不支持 IF NOT EXISTS for ALTER TABLE，需要先检查
		var columnExists int
		err := d.db.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('traders') 
			WHERE name='exchange_api_key_name'
		`).Scan(&columnExists)
		
		if err != nil {
			return fmt.Errorf("检查列是否存在失败: %w", err)
		}
		
		if columnExists == 0 {
			_, err := d.db.Exec(`ALTER TABLE traders ADD COLUMN exchange_api_key_name TEXT DEFAULT ''`)
			if err != nil {
				return fmt.Errorf("添加 exchange_api_key_name 列失败: %w", err)
			}
		}
	}
	
	return nil
}
