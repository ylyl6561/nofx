package config

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"log"
	"nofx/market"
	"os"
	"slices"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Database 配置数据库（仅支持PostgreSQL）
type Database struct {
	db *sql.DB
}

// convertQuery 将 ? 占位符转换为 PostgreSQL 的 $1, $2, $3... 格式
func (d *Database) convertQuery(query string) string {
	result := ""
	paramCount := 1
	for _, char := range query {
		if char == '?' {
			result += fmt.Sprintf("$%d", paramCount)
			paramCount++
		} else {
			result += string(char)
		}
	}
	return result
}

// isPostgreSQL 检查是否使用PostgreSQL（现在始终返回true）
func (d *Database) isPostgreSQL() bool {
	return true
}

// IsPostgreSQL 检查是否使用PostgreSQL（公开方法，现在始终返回true）
func (d *Database) IsPostgreSQL() bool {
	return true
}

// NewDatabase 创建PostgreSQL数据库连接
func NewDatabase(dbPath string) (*Database, error) {
	var db *sql.DB
	var err error
	
	// 优先使用 DATABASE_URL
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		db, err = sql.Open("postgres", dbURL)
		log.Printf("🐘 连接 PostgreSQL (DATABASE_URL): %s", maskURL(dbURL))
	} else {
		// 使用独立的环境变量构建连接字符串
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSL_MODE")
		
		if host == "" || user == "" || dbname == "" {
			return nil, fmt.Errorf("PostgreSQL 配置不完整，需要设置 DB_HOST, DB_USER, DB_NAME")
		}
		
		if port == "" {
			port = "5432"
		}
		if sslmode == "" {
			sslmode = "disable"
		}
		
		// 构建连接字符串
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
		
		db, err = sql.Open("postgres", connStr)
		log.Printf("🐘 连接 PostgreSQL: host=%s port=%s dbname=%s user=%s sslmode=%s",
			host, port, dbname, user, sslmode)
	}
	
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %w", err)
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	if err := database.initDefaultData(); err != nil {
		return nil, fmt.Errorf("初始化默认数据失败: %w", err)
	}

	return database, nil
}

// Query 执行查询并返回结果集（用于API层查询）
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// initDefaultData 初始化默认数据
func (d *Database) initDefaultData() error {
	// 初始化系统级 AI 模型模板
	systemAIModels := []struct {
		id, name, provider, description, defaultModelName, defaultAPIURL string
	}{
		{"deepseek", "DeepSeek", "deepseek", "DeepSeek AI 模型，性价比高", "deepseek-chat", "https://api.deepseek.com/v1"},
		{"qwen", "Qwen (通义千问)", "qwen", "阿里云通义千问模型", "qwen-plus", "https://dashscope.aliyuncs.com/compatible-mode/v1"},
	}

	for _, model := range systemAIModels {
		_, err := d.db.Exec(`
			INSERT INTO system_ai_models (id, name, provider, description, default_model_name, default_api_url) 
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING
		`, model.id, model.name, model.provider, model.description, model.defaultModelName, model.defaultAPIURL)
		if err != nil {
			return fmt.Errorf("初始化系统 AI 模型模板失败: %w", err)
		}
	}

	// 初始化系统级交易所模板
	systemExchanges := []struct {
		id, name, typ, description string
		supportsTestnet            bool
	}{
		{"binance", "Binance Futures", "cex", "币安合约交易所", true},
		{"okx", "OKX", "cex", "OKX 交易所", true},
		{"hyperliquid", "Hyperliquid", "dex", "Hyperliquid 去中心化交易所", true},
		{"aster", "Aster", "dex", "Aster 去中心化交易所", false},
	}

	for _, exchange := range systemExchanges {
		_, err := d.db.Exec(`
			INSERT INTO system_exchanges (id, name, type, description, supports_testnet) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
		`, exchange.id, exchange.name, exchange.typ, exchange.description, exchange.supportsTestnet)
		if err != nil {
			return fmt.Errorf("初始化系统交易所模板失败: %w", err)
		}
	}

	// 初始化AI模型（使用default用户）
	aiModels := []struct {
		id, name, provider string
	}{
		{"deepseek", "DeepSeek", "deepseek"},
		{"qwen", "Qwen (通义千问)", "qwen"},
		{"custom", "Custom OpenAI API", "custom"},
	}

	for _, model := range aiModels {
		_, err := d.db.Exec(`
			INSERT INTO ai_models (id, user_id, name, provider, enabled) 
			VALUES ($1, 'default', $2, $3, false)
			ON CONFLICT (id, user_id) DO NOTHING
		`, model.id, model.name, model.provider)
		if err != nil {
			return fmt.Errorf("初始化AI模型失败: %w", err)
		}
	}

	// 初始化交易所（使用default用户）
	exchanges := []struct {
		id, name, typ string
	}{
		{"binance", "Binance Futures", "cex"},
		{"okx", "OKX", "cex"},
		{"hyperliquid", "Hyperliquid", "dex"},
		{"aster", "Aster DEX", "dex"},
	}

	for _, exchange := range exchanges {
		_, err := d.db.Exec(`
			INSERT INTO exchanges (id, user_id, api_key_name, name, type, enabled) 
			VALUES ($1, 'default', '', $2, $3, false)
			ON CONFLICT (id, user_id, api_key_name) DO NOTHING
		`, exchange.id, exchange.name, exchange.typ)
		if err != nil {
			return fmt.Errorf("初始化交易所失败: %w", err)
		}
	}

	// 初始化系统配置 - 创建所有字段，设置默认值，后续由config.json同步更新
	systemConfigs := map[string]string{
		"beta_mode":                   "false",                                                                               // 默认关闭内测模式
		"api_server_port":             "8080",                                                                                // 默认API端口
		"use_default_coins":           "true",                                                                                // 默认使用内置币种列表
		"default_coins":               `["BTCUSDT","ETHUSDT","SOLUSDT","BNBUSDT","XRPUSDT","DOGEUSDT","ADAUSDT","HYPEUSDT"]`, // 默认币种列表（JSON格式）
		"max_daily_loss":              "10.0",                                                                                // 最大日损失百分比
		"max_drawdown":                "20.0",                                                                                // 最大回撤百分比
		"stop_trading_minutes":        "60",                                                                                  // 停止交易时间（分钟）
		"btc_eth_leverage":            "5",                                                                                   // BTC/ETH杠杆倍数
		"altcoin_leverage":            "5",                                                                                   // 山寨币杠杆倍数
		"jwt_secret":                  "",                                                                                    // JWT密钥，默认为空，由config.json或系统生成
		"decision_logs_retention_days": "10",                                                                                 // 决策日志保留天数
		"equity_history_retention_days": "90",                                                                                // 权益历史保留天数
	}

	for key, value := range systemConfigs {
		_, err := d.db.Exec(`
			INSERT INTO system_config (key, value) 
			VALUES ($1, $2)
			ON CONFLICT (key) DO NOTHING
		`, key, value)
		if err != nil {
			return fmt.Errorf("初始化系统配置失败 [%s]: %w", key, err)
		}
	}

	return nil
}

// migrateExchangesTable 迁移exchanges表支持多用户
func (d *Database) migrateExchangesTable() error {
	// 检查是否已经迁移过
	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='exchanges_new'
	`).Scan(&count)
	if err != nil {
		return err
	}

	// 如果已经迁移过，直接返回
	if count > 0 {
		return nil
	}

	log.Printf("🔄 开始迁移exchanges表...")

	// 创建新的exchanges表，使用复合主键
	_, err = d.db.Exec(`
		CREATE TABLE exchanges_new (
			id TEXT NOT NULL,
			user_id TEXT NOT NULL DEFAULT 'default',
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
			PRIMARY KEY (id, user_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("创建新exchanges表失败: %w", err)
	}

	// 复制数据到新表（明确指定列名，兼容有无 passphrase 的情况）
	_, err = d.db.Exec(`
		INSERT INTO exchanges_new (id, user_id, name, type, enabled, api_key, secret_key, passphrase, testnet, 
		                            hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key, created_at, updated_at)
		SELECT id, user_id, name, type, enabled, api_key, secret_key, 
		       COALESCE(passphrase, '') as passphrase, testnet,
		       COALESCE(hyperliquid_wallet_addr, '') as hyperliquid_wallet_addr,
		       COALESCE(aster_user, '') as aster_user,
		       COALESCE(aster_signer, '') as aster_signer,
		       COALESCE(aster_private_key, '') as aster_private_key,
		       created_at, updated_at
		FROM exchanges
	`)
	if err != nil {
		return fmt.Errorf("复制数据失败: %w", err)
	}

	// 删除旧表
	_, err = d.db.Exec(`DROP TABLE exchanges`)
	if err != nil {
		return fmt.Errorf("删除旧表失败: %w", err)
	}

	// 重命名新表
	_, err = d.db.Exec(`ALTER TABLE exchanges_new RENAME TO exchanges`)
	if err != nil {
		return fmt.Errorf("重命名表失败: %w", err)
	}

	// 重新创建触发器
	_, err = d.db.Exec(`
		CREATE TRIGGER IF NOT EXISTS update_exchanges_updated_at
			AFTER UPDATE ON exchanges
			BEGIN
				UPDATE exchanges SET updated_at = CURRENT_TIMESTAMP 
				WHERE id = NEW.id AND user_id = NEW.user_id;
			END
	`)
	if err != nil {
		return fmt.Errorf("创建触发器失败: %w", err)
	}

	log.Printf("✅ exchanges表迁移完成")
	return nil
}

// User 用户配置
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // 不返回到前端
	OTPSecret    string    `json:"-"` // 不返回到前端
	OTPVerified  bool      `json:"otp_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AIModelConfig AI模型配置
type AIModelConfig struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Provider        string    `json:"provider"`
	Enabled         bool      `json:"enabled"`
	APIKey          string    `json:"apiKey"`
	CustomAPIURL    string    `json:"customApiUrl"`
	CustomModelName string    `json:"customModelName"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Enabled    bool   `json:"enabled"`
	APIKeyName string `json:"apiKeyName"` // API密钥名称
	APIKey     string `json:"apiKey"`
	SecretKey  string `json:"secretKey"`
	Testnet    bool   `json:"testnet"`
	// OKX 特定字段
	Passphrase string `json:"passphrase"` // OKX API Passphrase
	// Hyperliquid 特定字段
	HyperliquidWalletAddr string `json:"hyperliquidWalletAddr"`
	// Aster 特定字段
	AsterUser       string    `json:"asterUser"`
	AsterSigner     string    `json:"asterSigner"`
	AsterPrivateKey string    `json:"asterPrivateKey"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TraderRecord 交易员配置（数据库实体）
type TraderRecord struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	Name                 string    `json:"name"`
	AIModelID            string    `json:"ai_model_id"`
	ExchangeID           string    `json:"exchange_id"`
	ExchangeAPIKeyName   string    `json:"exchange_api_key_name"`  // 交易所API密钥名称
	InitialBalance       float64   `json:"initial_balance"`
	ScanIntervalMinutes  int       `json:"scan_interval_minutes"`
	IsRunning            bool      `json:"is_running"`
	BTCETHLeverage       int       `json:"btc_eth_leverage"`       // BTC/ETH杠杆倍数
	AltcoinLeverage      int       `json:"altcoin_leverage"`       // 山寨币杠杆倍数
	TradingSymbols       string    `json:"trading_symbols"`        // 交易币种，逗号分隔
	UseCoinPool          bool      `json:"use_coin_pool"`          // 是否使用COIN POOL信号源
	UseOITop             bool      `json:"use_oi_top"`             // 是否使用OI TOP信号源
	CustomPrompt         string    `json:"custom_prompt"`          // 自定义交易策略prompt
	OverrideBasePrompt   bool      `json:"override_base_prompt"`   // 是否覆盖基础prompt
	SystemPromptTemplate string    `json:"system_prompt_template"` // 系统提示词模板名称
	IsCrossMargin        bool      `json:"is_cross_margin"`        // 是否为全仓模式（true=全仓，false=逐仓）
	TotalTokens          int64     `json:"total_tokens"`           // 累计使用的 token 数量
	IsDeleted            string    `json:"is_deleted"`             // 软删除标记 ('n'=未删除, 'y'=已删除)
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// UserSignalSource 用户信号源配置
type UserSignalSource struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	CoinPoolURL string    `json:"coin_pool_url"`
	OITopURL    string    `json:"oi_top_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GenerateOTPSecret 生成OTP密钥
func GenerateOTPSecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

// CreateUser 创建用户
func (d *Database) CreateUser(user *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, otp_secret, otp_verified)
		VALUES (?, ?, ?, ?, ?)`
	_, err := d.db.Exec(d.convertQuery(query), user.ID, user.Email, user.PasswordHash, user.OTPSecret, user.OTPVerified)
	return err
}

// EnsureAdminUser 确保admin用户存在（用于管理员模式）
func (d *Database) EnsureAdminUser() error {
	// 检查admin用户是否已存在
	var count int
	query := `SELECT COUNT(*) FROM users WHERE id = ?`
	err := d.db.QueryRow(d.convertQuery(query), "admin").Scan(&count)
	if err != nil {
		return err
	}

	// 如果已存在，直接返回
	if count > 0 {
		return nil
	}

	// 创建admin用户（密码为空，因为管理员模式下不需要密码）
	adminUser := &User{
		ID:           "admin",
		Email:        "admin@localhost",
		PasswordHash: "", // 管理员模式下不使用密码
		OTPSecret:    "",
		OTPVerified:  true,
	}

	return d.CreateUser(adminUser)
}

// GetUserByEmail 通过邮箱获取用户
func (d *Database) GetUserByEmail(email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password_hash, otp_secret, otp_verified, created_at, updated_at
		FROM users WHERE email = ?`
	err := d.db.QueryRow(d.convertQuery(query), email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.OTPSecret,
		&user.OTPVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 通过ID获取用户
func (d *Database) GetUserByID(userID string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password_hash, otp_secret, otp_verified, created_at, updated_at
		FROM users WHERE id = ?`
	err := d.db.QueryRow(d.convertQuery(query), userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.OTPSecret,
		&user.OTPVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers 获取所有用户ID列表
func (d *Database) GetAllUsers() ([]string, error) {
	rows, err := d.db.Query(`SELECT id FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

// UpdateUserOTPVerified 更新用户OTP验证状态
func (d *Database) UpdateUserOTPVerified(userID string, verified bool) error {
	_, err := d.db.Exec(d.convertQuery(`UPDATE users SET otp_verified = ? WHERE id = ?`), verified, userID)
	return err
}

// UpdateUserPassword 更新用户密码
func (d *Database) UpdateUserPassword(userID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`
	_, err := d.db.Exec(d.convertQuery(query), passwordHash, userID)
	return err
}

// GetAIModels 获取用户的AI模型配置
func (d *Database) GetAIModels(userID string) ([]*AIModelConfig, error) {
	query := `
		SELECT id, user_id, name, provider, enabled, api_key,
		       COALESCE(custom_api_url, '') as custom_api_url,
		       COALESCE(custom_model_name, '') as custom_model_name,
		       created_at, updated_at
		FROM ai_models WHERE user_id = ? ORDER BY id`
	rows, err := d.db.Query(d.convertQuery(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化为空切片而不是nil，确保JSON序列化为[]而不是null
	models := make([]*AIModelConfig, 0)
	for rows.Next() {
		var model AIModelConfig
		err := rows.Scan(
			&model.ID, &model.UserID, &model.Name, &model.Provider,
			&model.Enabled, &model.APIKey, &model.CustomAPIURL, &model.CustomModelName,
			&model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, &model)
	}

	return models, nil
}

// UpdateAIModel 更新AI模型配置，如果不存在则创建用户特定配置
func (d *Database) UpdateAIModel(userID, id string, enabled bool, apiKey, customAPIURL, customModelName string) error {
	// 先尝试精确匹配 ID（新版逻辑，支持多个相同 provider 的模型）
	var existingID string
	query1 := `SELECT id FROM ai_models WHERE user_id = ? AND id = ? LIMIT 1`
	err := d.db.QueryRow(d.convertQuery(query1), userID, id).Scan(&existingID)

	if err == nil {
		// 找到了现有配置（精确匹配 ID），更新它
		query2 := `
			UPDATE ai_models SET enabled = ?, api_key = ?, custom_api_url = ?, custom_model_name = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND user_id = ?`
		_, err = d.db.Exec(d.convertQuery(query2), enabled, apiKey, customAPIURL, customModelName, existingID, userID)
		return err
	}

	// ID 不存在，尝试兼容旧逻辑：将 id 作为 provider 查找
	provider := id
	query3 := `SELECT id FROM ai_models WHERE user_id = ? AND provider = ? LIMIT 1`
	err = d.db.QueryRow(d.convertQuery(query3), userID, provider).Scan(&existingID)

	if err == nil {
		// 找到了现有配置（通过 provider 匹配，兼容旧版），更新它
		log.Printf("⚠️  使用旧版 provider 匹配更新模型: %s -> %s", provider, existingID)
		query4 := `
			UPDATE ai_models SET enabled = ?, api_key = ?, custom_api_url = ?, custom_model_name = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND user_id = ?`
		_, err = d.db.Exec(d.convertQuery(query4), enabled, apiKey, customAPIURL, customModelName, existingID, userID)
		return err
	}

	// 没有找到任何现有配置，创建新的
	// 推断 provider（从 id 中提取，或者直接使用 id）
	if provider == id && (provider == "deepseek" || provider == "qwen") {
		// id 本身就是 provider
		provider = id
	} else {
		// 从 id 中提取 provider（假设格式是 userID_provider 或 timestamp_userID_provider）
		parts := strings.Split(id, "_")
		if len(parts) >= 2 {
			provider = parts[len(parts)-1] // 取最后一部分作为 provider
		} else {
			provider = id
		}
	}

	// 获取模型的基本信息
	var name string
	query5 := `SELECT name FROM ai_models WHERE provider = ? LIMIT 1`
	err = d.db.QueryRow(d.convertQuery(query5), provider).Scan(&name)
	if err != nil {
		// 如果找不到基本信息，使用默认值
		if provider == "deepseek" {
			name = "DeepSeek AI"
		} else if provider == "qwen" {
			name = "Qwen AI"
		} else {
			name = provider + " AI"
		}
	}

	// 如果传入的 ID 已经是完整格式（如 "admin_deepseek_custom1"），直接使用
	// 否则生成新的 ID
	newModelID := id
	if id == provider {
		// id 就是 provider，生成新的用户特定 ID
		newModelID = fmt.Sprintf("%s_%s", userID, provider)
	}

	log.Printf("✓ 创建新的 AI 模型配置: ID=%s, Provider=%s, Name=%s", newModelID, provider, name)
	query6 := `
		INSERT INTO ai_models (id, user_id, name, provider, enabled, api_key, custom_api_url, custom_model_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err = d.db.Exec(d.convertQuery(query6), newModelID, userID, name, provider, enabled, apiKey, customAPIURL, customModelName)

	return err
}

// DeleteAIModel 删除AI模型配置
func (d *Database) DeleteAIModel(userID, modelID string) error {
	log.Printf("🗑️  DeleteAIModel: userID=%s, modelID=%s", userID, modelID)
	
	query := `DELETE FROM ai_models WHERE id = ? AND user_id = ?`
	result, err := d.db.Exec(d.convertQuery(query), modelID, userID)
	if err != nil {
		log.Printf("❌ DeleteAIModel: 删除失败: %v", err)
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ DeleteAIModel: 获取影响行数失败: %v", err)
		return err
	}
	
	if rowsAffected == 0 {
		log.Printf("⚠️  DeleteAIModel: 未找到要删除的AI模型")
		return fmt.Errorf("AI模型不存在")
	}
	
	log.Printf("✓ DeleteAIModel: 成功删除 %d 条记录", rowsAffected)
	return nil
}

// GetExchanges 获取用户的交易所配置
func (d *Database) GetExchanges(userID string) ([]*ExchangeConfig, error) {
	query := `
		SELECT id, user_id, name, type, enabled, 
		       COALESCE(api_key_name, '') as api_key_name,
		       api_key, secret_key, testnet, 
		       COALESCE(passphrase, '') as passphrase,
		       COALESCE(hyperliquid_wallet_addr, '') as hyperliquid_wallet_addr,
		       COALESCE(aster_user, '') as aster_user,
		       COALESCE(aster_signer, '') as aster_signer,
		       COALESCE(aster_private_key, '') as aster_private_key,
		       created_at, updated_at 
		FROM exchanges WHERE user_id = ? ORDER BY id`
	rows, err := d.db.Query(d.convertQuery(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化为空切片而不是nil，确保JSON序列化为[]而不是null
	exchanges := make([]*ExchangeConfig, 0)
	for rows.Next() {
		var exchange ExchangeConfig
		err := rows.Scan(
			&exchange.ID, &exchange.UserID, &exchange.Name, &exchange.Type,
			&exchange.Enabled, &exchange.APIKeyName, &exchange.APIKey, &exchange.SecretKey, &exchange.Testnet,
			&exchange.Passphrase,
			&exchange.HyperliquidWalletAddr, &exchange.AsterUser,
			&exchange.AsterSigner, &exchange.AsterPrivateKey,
			&exchange.CreatedAt, &exchange.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exchanges = append(exchanges, &exchange)
	}

	return exchanges, nil
}

func (d *Database) UpdateExchange(userID, id string, enabled bool, apiKeyName, apiKey, secretKey, passphrase string, testnet bool, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey string) error {
	log.Printf("🔧 UpdateExchange: userID=%s, id=%s, enabled=%v, apiKeyName=%s", userID, id, enabled, apiKeyName)

	// 首先尝试更新现有的用户配置
	// 使用静态的 UPDATE 语句，避免动态拼接导致的语法错误
	// 所有字段都会被更新，空字符串会覆盖旧值
	query := `
		UPDATE exchanges SET 
			enabled = ?, 
			api_key = ?, 
			secret_key = ?, 
			passphrase = ?, 
			testnet = ?, 
			hyperliquid_wallet_addr = ?, 
			aster_user = ?, 
			aster_signer = ?, 
			aster_private_key = ?, 
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND api_key_name = ?`
	result, err := d.db.Exec(d.convertQuery(query), enabled, apiKey, secretKey, passphrase, testnet, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey, id, userID, apiKeyName)
	if err != nil {
		log.Printf("❌ UpdateExchange: 更新失败: %v", err)
		return err
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ UpdateExchange: 获取影响行数失败: %v", err)
		return err
	}

	log.Printf("📊 UpdateExchange: 影响行数 = %d", rowsAffected)

	// 如果没有行被更新，说明用户没有这个交易所的配置，需要创建
	if rowsAffected == 0 {
		log.Printf("💡 UpdateExchange: 没有现有记录，创建新记录")

		// 根据交易所ID确定基本信息
		var name, typ string
		if id == "binance" {
			name = "Binance Futures"
			typ = "cex"
		} else if id == "hyperliquid" {
			name = "Hyperliquid"
			typ = "dex"
		} else if id == "aster" {
			name = "Aster DEX"
			typ = "dex"
		} else {
			name = id + " Exchange"
			typ = "cex"
		}

		log.Printf("🆕 UpdateExchange: 创建新记录 ID=%s, name=%s, type=%s", id, name, typ)

		// 创建用户特定的配置，使用原始的交易所ID
		insertQuery := `
			INSERT INTO exchanges (id, user_id, name, type, enabled, api_key_name, api_key, secret_key, passphrase, testnet, 
			                       hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = d.db.Exec(d.convertQuery(insertQuery), id, userID, name, typ, enabled, apiKeyName, apiKey, secretKey, passphrase, testnet, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey)

		if err != nil {
			log.Printf("❌ UpdateExchange: 创建记录失败: %v", err)
		} else {
			log.Printf("✅ UpdateExchange: 创建记录成功")
		}
		return err
	}

	log.Printf("✅ UpdateExchange: 更新现有记录成功")
	return nil
}

// DeleteExchange 删除交易所配置
func (d *Database) DeleteExchange(userID, exchangeID, apiKeyName string) error {
	log.Printf("🗑️  DeleteExchange: userID=%s, exchangeID=%s, apiKeyName=%s", userID, exchangeID, apiKeyName)
	
	query := `DELETE FROM exchanges WHERE id = ? AND user_id = ? AND api_key_name = ?`
	result, err := d.db.Exec(d.convertQuery(query), exchangeID, userID, apiKeyName)
	if err != nil {
		log.Printf("❌ DeleteExchange: 删除失败: %v", err)
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ DeleteExchange: 获取影响行数失败: %v", err)
		return err
	}
	
	if rowsAffected == 0 {
		log.Printf("⚠️  DeleteExchange: 未找到要删除的记录")
		return fmt.Errorf("交易所配置不存在")
	}
	
	log.Printf("✅ DeleteExchange: 删除成功，影响行数 = %d", rowsAffected)
	return nil
}

// CreateAIModel 创建AI模型配置
func (d *Database) CreateAIModel(userID, id, name, provider string, enabled bool, apiKey, customAPIURL string) error {
	_, err := d.db.Exec(`
		INSERT INTO ai_models (id, user_id, name, provider, enabled, api_key, custom_api_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id, user_id) DO NOTHING
	`, id, userID, name, provider, enabled, apiKey, customAPIURL)
	return err
}

// CreateExchange 创建交易所配置
func (d *Database) CreateExchange(userID, id, apiKeyName, name, typ string, enabled bool, apiKey, secretKey string, testnet bool, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey string) error {
	_, err := d.db.Exec(`
		INSERT INTO exchanges (id, user_id, api_key_name, name, type, enabled, api_key, secret_key, testnet, hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id, user_id, api_key_name) DO NOTHING
	`, id, userID, apiKeyName, name, typ, enabled, apiKey, secretKey, testnet, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey)
	return err
}

// CreateTrader 创建交易员
func (d *Database) CreateTrader(trader *TraderRecord) error {
	query := `
		INSERT INTO traders (id, user_id, name, ai_model_id, exchange_id, exchange_api_key_name, initial_balance, scan_interval_minutes, is_running, btc_eth_leverage, altcoin_leverage, trading_symbols, use_coin_pool, use_oi_top, custom_prompt, override_base_prompt, system_prompt_template, is_cross_margin)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(d.convertQuery(query), trader.ID, trader.UserID, trader.Name, trader.AIModelID, trader.ExchangeID, trader.ExchangeAPIKeyName, trader.InitialBalance, trader.ScanIntervalMinutes, trader.IsRunning, trader.BTCETHLeverage, trader.AltcoinLeverage, trader.TradingSymbols, trader.UseCoinPool, trader.UseOITop, trader.CustomPrompt, trader.OverrideBasePrompt, trader.SystemPromptTemplate, trader.IsCrossMargin)
	return err
}

// GetTraders 获取用户的交易员（过滤已删除）
func (d *Database) GetTraders(userID string) ([]*TraderRecord, error) {
	query := `
		SELECT id, user_id, name, ai_model_id, exchange_id, COALESCE(exchange_api_key_name, '') as exchange_api_key_name,
		       initial_balance, scan_interval_minutes, is_running,
		       COALESCE(btc_eth_leverage, 5) as btc_eth_leverage, COALESCE(altcoin_leverage, 5) as altcoin_leverage,
		       COALESCE(trading_symbols, '') as trading_symbols,
		       COALESCE(use_coin_pool, false) as use_coin_pool, COALESCE(use_oi_top, false) as use_oi_top,
		       COALESCE(custom_prompt, '') as custom_prompt, COALESCE(override_base_prompt, false) as override_base_prompt,
		       COALESCE(system_prompt_template, 'default') as system_prompt_template,
		       COALESCE(is_cross_margin, true) as is_cross_margin,
		       COALESCE(total_tokens, 0) as total_tokens, COALESCE(is_deleted, 'n') as is_deleted,
		       created_at, updated_at
		FROM traders WHERE user_id = ? AND COALESCE(is_deleted, 'n') = 'n' ORDER BY created_at DESC`
	
	rows, err := d.db.Query(d.convertQuery(query), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traders []*TraderRecord
	for rows.Next() {
		var trader TraderRecord
		err := rows.Scan(
			&trader.ID, &trader.UserID, &trader.Name, &trader.AIModelID, &trader.ExchangeID, &trader.ExchangeAPIKeyName,
			&trader.InitialBalance, &trader.ScanIntervalMinutes, &trader.IsRunning,
			&trader.BTCETHLeverage, &trader.AltcoinLeverage, &trader.TradingSymbols,
			&trader.UseCoinPool, &trader.UseOITop,
			&trader.CustomPrompt, &trader.OverrideBasePrompt, &trader.SystemPromptTemplate,
			&trader.IsCrossMargin,
			&trader.TotalTokens, &trader.IsDeleted,
			&trader.CreatedAt, &trader.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		traders = append(traders, &trader)
	}

	return traders, nil
}

// UpdateTraderStatus 更新交易员状态
func (d *Database) UpdateTraderStatus(userID, id string, isRunning bool) error {
	_, err := d.db.Exec(d.convertQuery(`UPDATE traders SET is_running = ? WHERE id = ? AND user_id = ?`), isRunning, id, userID)
	return err
}

// UpdateTraderTokens 累加 trader 的 token 使用量
func (d *Database) UpdateTraderTokens(traderID string, tokens int) error {
	query := d.convertQuery("UPDATE traders SET total_tokens = total_tokens + ? WHERE id = ?")
	_, err := d.db.Exec(query, tokens, traderID)
	return err
}

// UpdateUserTokens 累加 user 的 token 使用量
func (d *Database) UpdateUserTokens(userID string, tokens int) error {
	query := d.convertQuery("UPDATE users SET total_tokens = total_tokens + ? WHERE id = ?")
	_, err := d.db.Exec(query, tokens, userID)
	return err
}

// GetUserTotalTokens 获取用户的总 token 使用量
func (d *Database) GetUserTotalTokens(userID string) (int64, error) {
	var totalTokens int64
	query := d.convertQuery("SELECT COALESCE(total_tokens, 0) FROM users WHERE id = ?")
	err := d.db.QueryRow(query, userID).Scan(&totalTokens)
	return totalTokens, err
}

// SoftDeleteTrader 软删除交易员（设置 is_deleted = 'y'）
func (d *Database) SoftDeleteTrader(userID, traderID string) error {
	query := d.convertQuery("UPDATE traders SET is_deleted = 'y', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?")
	_, err := d.db.Exec(query, traderID, userID)
	return err
}

// UpdateTrader 更新交易员配置
func (d *Database) UpdateTrader(trader *TraderRecord) error {
	query := `
		UPDATE traders SET
			name = ?, ai_model_id = ?, exchange_id = ?, initial_balance = ?,
			scan_interval_minutes = ?, btc_eth_leverage = ?, altcoin_leverage = ?,
			trading_symbols = ?, custom_prompt = ?, override_base_prompt = ?,
			system_prompt_template = ?, is_cross_margin = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ?`
	_, err := d.db.Exec(d.convertQuery(query), trader.Name, trader.AIModelID, trader.ExchangeID, trader.InitialBalance,
		trader.ScanIntervalMinutes, trader.BTCETHLeverage, trader.AltcoinLeverage,
		trader.TradingSymbols, trader.CustomPrompt, trader.OverrideBasePrompt,
		trader.SystemPromptTemplate, trader.IsCrossMargin, trader.ID, trader.UserID)
	return err
}

// UpdateTraderCustomPrompt 更新交易员自定义Prompt
func (d *Database) UpdateTraderCustomPrompt(userID, id string, customPrompt string, overrideBase bool) error {
	_, err := d.db.Exec(d.convertQuery(`UPDATE traders SET custom_prompt = ?, override_base_prompt = ? WHERE id = ? AND user_id = ?`), customPrompt, overrideBase, id, userID)
	return err
}

// UpdateTraderInitialBalance 更新交易员初始余额（用于自动同步交易所实际余额）
func (d *Database) UpdateTraderInitialBalance(userID, id string, newBalance float64) error {
	_, err := d.db.Exec(d.convertQuery(`UPDATE traders SET initial_balance = ? WHERE id = ? AND user_id = ?`), newBalance, id, userID)
	return err
}

// DeleteTrader 删除交易员
func (d *Database) DeleteTrader(userID, id string) error {
	_, err := d.db.Exec(d.convertQuery(`DELETE FROM traders WHERE id = ? AND user_id = ?`), id, userID)
	return err
}

// GetTraderConfig 获取交易员完整配置（包含AI模型和交易所信息）
func (d *Database) GetTraderConfig(userID, traderID string) (*TraderRecord, *AIModelConfig, *ExchangeConfig, error) {
	var trader TraderRecord
	var aiModel AIModelConfig
	var exchange ExchangeConfig

	query := `
		SELECT
			t.id, t.user_id, t.name, t.ai_model_id, t.exchange_id, COALESCE(t.exchange_api_key_name, '') as exchange_api_key_name,
			t.initial_balance, t.scan_interval_minutes, t.is_running,
			COALESCE(t.btc_eth_leverage, 5) as btc_eth_leverage,
			COALESCE(t.altcoin_leverage, 5) as altcoin_leverage,
			COALESCE(t.trading_symbols, '') as trading_symbols,
			COALESCE(t.use_coin_pool, false) as use_coin_pool,
			COALESCE(t.use_oi_top, false) as use_oi_top,
			COALESCE(t.custom_prompt, '') as custom_prompt,
			COALESCE(t.override_base_prompt, false) as override_base_prompt,
			COALESCE(t.system_prompt_template, 'default') as system_prompt_template,
			COALESCE(t.is_cross_margin, true) as is_cross_margin,
			COALESCE(t.total_tokens, 0) as total_tokens,
			COALESCE(t.is_deleted, 'n') as is_deleted,
			t.created_at, t.updated_at,
			a.id, a.user_id, a.name, a.provider, a.enabled, a.api_key,
			COALESCE(a.custom_api_url, '') as custom_api_url,
			COALESCE(a.custom_model_name, '') as custom_model_name,
			a.created_at, a.updated_at,
			e.id, e.user_id, COALESCE(e.api_key_name, '') as api_key_name, e.name, e.type, e.enabled, e.api_key, e.secret_key, e.testnet,
			COALESCE(e.hyperliquid_wallet_addr, '') as hyperliquid_wallet_addr,
			COALESCE(e.aster_user, '') as aster_user,
			COALESCE(e.aster_signer, '') as aster_signer,
			COALESCE(e.aster_private_key, '') as aster_private_key,
			COALESCE(e.passphrase, '') as passphrase,
			e.created_at, e.updated_at
		FROM traders t
		JOIN ai_models a ON t.ai_model_id = a.id AND t.user_id = a.user_id
		LEFT JOIN exchanges e ON t.exchange_id = e.id AND t.user_id = e.user_id 
			AND (
				-- 精确匹配 api_key_name
				(COALESCE(t.exchange_api_key_name, '') != '' AND COALESCE(t.exchange_api_key_name, '') = COALESCE(e.api_key_name, ''))
				-- 或者 trader 的 api_key_name 为空时，匹配任意启用的交易所
				OR (COALESCE(t.exchange_api_key_name, '') = '' AND e.enabled = true)
			)
		WHERE t.id = ? AND t.user_id = ? AND COALESCE(t.is_deleted, 'n') = 'n'
		ORDER BY 
			-- 优先选择启用的交易所
			CASE WHEN e.enabled = true THEN 0 ELSE 1 END,
			-- 然后按创建时间倒序
			e.created_at DESC
		LIMIT 1`
	
	err := d.db.QueryRow(d.convertQuery(query), traderID, userID).Scan(
		&trader.ID, &trader.UserID, &trader.Name, &trader.AIModelID, &trader.ExchangeID, &trader.ExchangeAPIKeyName,
		&trader.InitialBalance, &trader.ScanIntervalMinutes, &trader.IsRunning,
		&trader.BTCETHLeverage, &trader.AltcoinLeverage, &trader.TradingSymbols,
		&trader.UseCoinPool, &trader.UseOITop,
		&trader.CustomPrompt, &trader.OverrideBasePrompt, &trader.SystemPromptTemplate,
		&trader.IsCrossMargin,
		&trader.TotalTokens, &trader.IsDeleted,
		&trader.CreatedAt, &trader.UpdatedAt,
		&aiModel.ID, &aiModel.UserID, &aiModel.Name, &aiModel.Provider, &aiModel.Enabled, &aiModel.APIKey,
		&aiModel.CustomAPIURL, &aiModel.CustomModelName,
		&aiModel.CreatedAt, &aiModel.UpdatedAt,
		&exchange.ID, &exchange.UserID, &exchange.APIKeyName, &exchange.Name, &exchange.Type, &exchange.Enabled,
		&exchange.APIKey, &exchange.SecretKey, &exchange.Testnet,
		&exchange.HyperliquidWalletAddr, &exchange.AsterUser, &exchange.AsterSigner, &exchange.AsterPrivateKey,
		&exchange.Passphrase,
		&exchange.CreatedAt, &exchange.UpdatedAt,
	)

	if err != nil {
		return nil, nil, nil, err
	}

	return &trader, &aiModel, &exchange, nil
}

// GetSystemConfig 获取系统配置
func (d *Database) GetSystemConfig(key string) (string, error) {
	var value string
	err := d.db.QueryRow(`SELECT value FROM system_config WHERE key = $1`, key).Scan(&value)
	return value, err
}

// SetSystemConfig 设置系统配置
func (d *Database) SetSystemConfig(key, value string) error {
	_, err := d.db.Exec(`
		INSERT INTO system_config (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}

// CreateUserSignalSource 创建用户信号源配置
func (d *Database) CreateUserSignalSource(userID, coinPoolURL, oiTopURL string) error {
	_, err := d.db.Exec(`
		INSERT INTO user_signal_sources (user_id, coin_pool_url, oi_top_url, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET 
			coin_pool_url = EXCLUDED.coin_pool_url,
			oi_top_url = EXCLUDED.oi_top_url,
			updated_at = CURRENT_TIMESTAMP
	`, userID, coinPoolURL, oiTopURL)
	return err
}

// GetUserSignalSource 获取用户信号源配置
func (d *Database) GetUserSignalSource(userID string) (*UserSignalSource, error) {
	var source UserSignalSource
	query := `
		SELECT id, user_id, coin_pool_url, oi_top_url, created_at, updated_at
		FROM user_signal_sources WHERE user_id = ?`
	err := d.db.QueryRow(d.convertQuery(query), userID).Scan(
		&source.ID, &source.UserID, &source.CoinPoolURL, &source.OITopURL,
		&source.CreatedAt, &source.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// UpdateUserSignalSource 更新用户信号源配置
func (d *Database) UpdateUserSignalSource(userID, coinPoolURL, oiTopURL string) error {
	query := `
		UPDATE user_signal_sources SET coin_pool_url = ?, oi_top_url = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?`
	_, err := d.db.Exec(d.convertQuery(query), coinPoolURL, oiTopURL, userID)
	return err
}

// GetCustomCoins 获取所有交易员自定义币种 / Get all trader-customized currencies
func (d *Database) GetCustomCoins() []string {
	var symbol string
	var symbols []string
	_ = d.db.QueryRow(`
		SELECT GROUP_CONCAT(custom_coins , ',') as symbol
		FROM main.traders where custom_coins != ''
	`).Scan(&symbol)
	// 检测用户是否未配置币种 - 兼容性
	if symbol == "" {
		symbolJSON, _ := d.GetSystemConfig("default_coins")
		if err := json.Unmarshal([]byte(symbolJSON), &symbols); err != nil {
			log.Printf("⚠️  解析default_coins配置失败: %v，使用硬编码默认值", err)
			symbols = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT"}
		}
	}
	// filter Symbol
	for _, s := range strings.Split(symbol, ",") {
		if s == "" {
			continue
		}
		coin := market.Normalize(s)
		if !slices.Contains(symbols, coin) {
			symbols = append(symbols, coin)
		}
	}
	return symbols
}

// IsUsingTestnet 检查是否使用测试网（从第一个启用的交易所读取）
func (d *Database) IsUsingTestnet() bool {
	var testnet bool
	query := `
		SELECT testnet FROM exchanges 
		WHERE enabled = true 
		LIMIT 1`
	err := d.db.QueryRow(d.convertQuery(query)).Scan(&testnet)
	
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("ℹ️  暂无启用的交易所配置，默认使用实盘模式")
		} else {
			log.Printf("⚠️  检查测试网配置失败: %v，默认使用实盘", err)
		}
		return false
	}
	
	if testnet {
		log.Printf("🧪 系统使用测试网模式")
	} else {
		log.Printf("💰 系统使用实盘模式")
	}
	
	return testnet
}

// maskURL 隐藏数据库URL中的敏感信息
func maskURL(dbURL string) string {
	if strings.Contains(dbURL, "@") {
		parts := strings.Split(dbURL, "@")
		if len(parts) >= 2 {
			// 保留协议部分，隐藏用户名密码
			protocol := strings.Split(parts[0], "://")
			if len(protocol) >= 2 {
				return protocol[0] + "://***:***@" + parts[1]
			}
		}
	}
	return "***"
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// LoadBetaCodesFromFile 从文件加载内测码到数据库
func (d *Database) LoadBetaCodesFromFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取内测码文件失败: %w", err)
	}

	// 按行分割内测码
	lines := strings.Split(string(content), "\n")
	var codes []string
	for _, line := range lines {
		code := strings.TrimSpace(line)
		if code != "" && !strings.HasPrefix(code, "#") {
			codes = append(codes, code)
		}
	}

	// 批量插入内测码
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	var stmt *sql.Stmt
	if d.isPostgreSQL() {
		// PostgreSQL 使用 ON CONFLICT
		stmt, err = tx.Prepare(`INSERT INTO beta_codes (code) VALUES ($1) ON CONFLICT (code) DO NOTHING`)
	} else {
		// SQLite 使用 INSERT OR IGNORE
		stmt, err = tx.Prepare(`INSERT OR IGNORE INTO beta_codes (code) VALUES (?)`)
	}
	if err != nil {
		return fmt.Errorf("准备语句失败: %w", err)
	}
	defer stmt.Close()

	insertedCount := 0
	for _, code := range codes {
		result, err := stmt.Exec(code)
		if err != nil {
			log.Printf("插入内测码 %s 失败: %v", code, err)
			continue
		}

		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
			insertedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("✅ 成功加载 %d 个内测码到数据库 (总计 %d 个)", insertedCount, len(codes))
	return nil
}

// ValidateBetaCode 验证内测码是否有效且未使用
func (d *Database) ValidateBetaCode(code string) (bool, error) {
	var used bool
	query := `SELECT used FROM beta_codes WHERE code = ?`
	err := d.db.QueryRow(d.convertQuery(query), code).Scan(&used)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // 内测码不存在
		}
		return false, err
	}
	return !used, nil // 内测码存在且未使用
}

// UseBetaCode 使用内测码（标记为已使用）
func (d *Database) UseBetaCode(code, userEmail string) error {
	query := `
		UPDATE beta_codes SET used = 1, used_by = ?, used_at = CURRENT_TIMESTAMP 
		WHERE code = ? AND used = 0`
	result, err := d.db.Exec(d.convertQuery(query), userEmail, code)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("内测码无效或已被使用")
	}

	return nil
}

// GetBetaCodeStats 获取内测码统计信息
func (d *Database) GetBetaCodeStats() (total, used int, err error) {
	err = d.db.QueryRow(`SELECT COUNT(*) FROM beta_codes`).Scan(&total)
	if err != nil {
		return 0, 0, err
	}

	query2 := `SELECT COUNT(*) FROM beta_codes WHERE used = 1`
	err = d.db.QueryRow(d.convertQuery(query2)).Scan(&used)
	if err != nil {
		return 0, 0, err
	}

	return total, used, nil
}
