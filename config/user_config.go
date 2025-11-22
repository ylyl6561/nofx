package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// UserTradingConfig 用户交易配置
type UserTradingConfig struct {
	UserID              string    `json:"user_id"`
	BTCETHLeverage      *int      `json:"btc_eth_leverage"`       // nil表示使用系统默认
	AltcoinLeverage     *int      `json:"altcoin_leverage"`       // nil表示使用系统默认
	MaxDailyLoss        *float64  `json:"max_daily_loss"`         // nil表示使用系统默认
	MaxDrawdown         *float64  `json:"max_drawdown"`           // nil表示使用系统默认
	StopTradingMinutes  *int      `json:"stop_trading_minutes"`   // nil表示使用系统默认
	UseDefaultCoins     bool      `json:"use_default_coins"`
	CustomCoins         []string  `json:"custom_coins"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GetUserTradingConfig 获取用户交易配置（合并系统默认值）
func (d *Database) GetUserTradingConfig(userID string) (*UserTradingConfig, error) {
	// 查询用户配置
	var config UserTradingConfig
	var btcEthLeverage, altcoinLeverage, stopTradingMinutes sql.NullInt64
	var maxDailyLoss, maxDrawdown sql.NullFloat64
	var customCoinsJSON string
	
	query := d.convertQuery(`
		SELECT user_id, btc_eth_leverage, altcoin_leverage, max_daily_loss, max_drawdown, 
		       stop_trading_minutes, use_default_coins, custom_coins, created_at, updated_at
		FROM user_trading_config
		WHERE user_id = ?
	`)
	
	err := d.db.QueryRow(query, userID).Scan(
		&config.UserID, &btcEthLeverage, &altcoinLeverage, &maxDailyLoss, &maxDrawdown,
		&stopTradingMinutes, &config.UseDefaultCoins, &customCoinsJSON, &config.CreatedAt, &config.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// 用户未配置，返回系统默认值
		return d.getSystemDefaultConfig(userID)
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户配置失败: %w", err)
	}
	
	// 处理可空字段
	if btcEthLeverage.Valid {
		val := int(btcEthLeverage.Int64)
		config.BTCETHLeverage = &val
	}
	if altcoinLeverage.Valid {
		val := int(altcoinLeverage.Int64)
		config.AltcoinLeverage = &val
	}
	if maxDailyLoss.Valid {
		config.MaxDailyLoss = &maxDailyLoss.Float64
	}
	if maxDrawdown.Valid {
		config.MaxDrawdown = &maxDrawdown.Float64
	}
	if stopTradingMinutes.Valid {
		val := int(stopTradingMinutes.Int64)
		config.StopTradingMinutes = &val
	}
	
	// 解析自定义币种
	if customCoinsJSON != "" {
		if err := json.Unmarshal([]byte(customCoinsJSON), &config.CustomCoins); err != nil {
			config.CustomCoins = []string{}
		}
	}
	
	// 合并系统默认值
	return d.mergeWithSystemDefaults(&config)
}

// getSystemDefaultConfig 获取系统默认配置
func (d *Database) getSystemDefaultConfig(userID string) (*UserTradingConfig, error) {
	config := &UserTradingConfig{
		UserID:          userID,
		UseDefaultCoins: true,
		CustomCoins:     []string{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	return d.mergeWithSystemDefaults(config)
}

// mergeWithSystemDefaults 合并系统默认值
func (d *Database) mergeWithSystemDefaults(config *UserTradingConfig) (*UserTradingConfig, error) {
	// 如果用户配置为空，使用系统默认值
	if config.BTCETHLeverage == nil {
		if val, err := d.GetSystemConfig("btc_eth_leverage"); err == nil {
			if intVal, err := strconv.Atoi(val); err == nil {
				config.BTCETHLeverage = &intVal
			}
		}
	}
	
	if config.AltcoinLeverage == nil {
		if val, err := d.GetSystemConfig("altcoin_leverage"); err == nil {
			if intVal, err := strconv.Atoi(val); err == nil {
				config.AltcoinLeverage = &intVal
			}
		}
	}
	
	if config.MaxDailyLoss == nil {
		if val, err := d.GetSystemConfig("max_daily_loss"); err == nil {
			if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				config.MaxDailyLoss = &floatVal
			}
		}
	}
	
	if config.MaxDrawdown == nil {
		if val, err := d.GetSystemConfig("max_drawdown"); err == nil {
			if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				config.MaxDrawdown = &floatVal
			}
		}
	}
	
	if config.StopTradingMinutes == nil {
		if val, err := d.GetSystemConfig("stop_trading_minutes"); err == nil {
			if intVal, err := strconv.Atoi(val); err == nil {
				config.StopTradingMinutes = &intVal
			}
		}
	}
	
	return config, nil
}

// SaveUserTradingConfig 保存用户交易配置
func (d *Database) SaveUserTradingConfig(config *UserTradingConfig) error {
	// 序列化自定义币种
	customCoinsJSON, err := json.Marshal(config.CustomCoins)
	if err != nil {
		return fmt.Errorf("序列化自定义币种失败: %w", err)
	}
	
	if d.isPostgreSQL() {
		_, err = d.db.Exec(`
			INSERT INTO user_trading_config 
			(user_id, btc_eth_leverage, altcoin_leverage, max_daily_loss, max_drawdown, 
			 stop_trading_minutes, use_default_coins, custom_coins)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (user_id) DO UPDATE SET
				btc_eth_leverage = EXCLUDED.btc_eth_leverage,
				altcoin_leverage = EXCLUDED.altcoin_leverage,
				max_daily_loss = EXCLUDED.max_daily_loss,
				max_drawdown = EXCLUDED.max_drawdown,
				stop_trading_minutes = EXCLUDED.stop_trading_minutes,
				use_default_coins = EXCLUDED.use_default_coins,
				custom_coins = EXCLUDED.custom_coins,
				updated_at = CURRENT_TIMESTAMP
		`, config.UserID, config.BTCETHLeverage, config.AltcoinLeverage, config.MaxDailyLoss,
			config.MaxDrawdown, config.StopTradingMinutes, config.UseDefaultCoins, string(customCoinsJSON))
	} else {
		_, err = d.db.Exec(d.convertQuery(`
			INSERT INTO user_trading_config 
			(user_id, btc_eth_leverage, altcoin_leverage, max_daily_loss, max_drawdown, 
			 stop_trading_minutes, use_default_coins, custom_coins, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id) DO UPDATE SET
				btc_eth_leverage = EXCLUDED.btc_eth_leverage,
				altcoin_leverage = EXCLUDED.altcoin_leverage,
				max_daily_loss = EXCLUDED.max_daily_loss,
				max_drawdown = EXCLUDED.max_drawdown,
				stop_trading_minutes = EXCLUDED.stop_trading_minutes,
				use_default_coins = EXCLUDED.use_default_coins,
				custom_coins = EXCLUDED.custom_coins,
				updated_at = CURRENT_TIMESTAMP
		`), config.UserID, config.BTCETHLeverage, config.AltcoinLeverage, config.MaxDailyLoss,
			config.MaxDrawdown, config.StopTradingMinutes, config.UseDefaultCoins, string(customCoinsJSON))
	}
	
	if err != nil {
		return fmt.Errorf("保存用户配置失败: %w", err)
	}
	
	return nil
}

// DeleteUserTradingConfig 删除用户交易配置（恢复使用系统默认值）
func (d *Database) DeleteUserTradingConfig(userID string) error {
	query := d.convertQuery(`DELETE FROM user_trading_config WHERE user_id = ?`)
	_, err := d.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("删除用户配置失败: %w", err)
	}
	return nil
}
