package config

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// APIKey API密钥结构
type APIKey struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	KeyHash      string     `json:"-"` // 不返回给前端
	KeyPrefix    string     `json:"key_prefix"`
	Name         string     `json:"name"`
	Enabled      bool       `json:"enabled"`
	RateLimit    int        `json:"rate_limit"`
	UsageCount   int        `json:"usage_count"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// GenerateAPIKey 生成新的API Key
func GenerateAPIKey() (apiKey string, keyHash string, keyPrefix string, err error) {
	// 生成32字节随机数
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("生成随机数失败: %w", err)
	}

	// 生成API Key（格式：aitrader_xxxxx）
	apiKey = "aitrader_" + base64.URLEncoding.EncodeToString(b)

	// 生成哈希值（存储在数据库）
	hash := sha256.Sum256([]byte(apiKey))
	keyHash = base64.URLEncoding.EncodeToString(hash[:])

	// 生成前缀（用于显示）
	keyPrefix = apiKey[:12] + "..." + apiKey[len(apiKey)-4:]

	return apiKey, keyHash, keyPrefix, nil
}

// CreateAPIKey 创建新的API Key
func (d *Database) CreateAPIKey(userID, name string, rateLimit int) (string, *APIKey, error) {
	// 生成API Key
	apiKey, keyHash, keyPrefix, err := GenerateAPIKey()
	if err != nil {
		return "", nil, err
	}

	// 生成ID
	id := uuid.New().String()

	// 插入数据库
	_, err = d.db.Exec(`
		INSERT INTO api_keys (id, user_id, key_hash, key_prefix, name, rate_limit)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, userID, keyHash, keyPrefix, name, rateLimit)

	if err != nil {
		return "", nil, fmt.Errorf("创建API Key失败: %w", err)
	}

	// 返回API Key信息
	keyInfo := &APIKey{
		ID:         id,
		UserID:     userID,
		KeyPrefix:  keyPrefix,
		Name:       name,
		Enabled:    true,
		RateLimit:  rateLimit,
		UsageCount: 0,
		CreatedAt:  time.Now(),
	}

	return apiKey, keyInfo, nil
}

// ValidateAPIKey 验证API Key
func (d *Database) ValidateAPIKey(apiKey string) (*APIKey, error) {
	// 计算哈希
	hash := sha256.Sum256([]byte(apiKey))
	keyHash := base64.URLEncoding.EncodeToString(hash[:])

	// 查询数据库
	var key APIKey
	var lastUsedAt sql.NullTime
	var expiresAt sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, user_id, key_prefix, name, enabled, rate_limit, usage_count, 
		       last_used_at, created_at, expires_at
		FROM api_keys
		WHERE key_hash = ?
	`, keyHash).Scan(
		&key.ID, &key.UserID, &key.KeyPrefix, &key.Name, &key.Enabled,
		&key.RateLimit, &key.UsageCount, &lastUsedAt, &key.CreatedAt, &expiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, err
	}

	// 处理可空字段
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}

	// 检查是否启用
	if !key.Enabled {
		return nil, fmt.Errorf("API key is disabled")
	}

	// 检查是否过期
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// 检查速率限制（简单版本：每月限制）
	if key.UsageCount >= key.RateLimit {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	return &key, nil
}

// IncrementAPIUsage 增加API使用次数
func (d *Database) IncrementAPIUsage(keyID string) error {
	_, err := d.db.Exec(`
		UPDATE api_keys 
		SET usage_count = usage_count + 1, last_used_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, keyID)
	return err
}

// LogAPIUsage 记录API使用日志
func (d *Database) LogAPIUsage(keyID, userID, endpoint, method string, statusCode, responseTimeMs int) error {
	_, err := d.db.Exec(`
		INSERT INTO api_usage_logs (api_key_id, user_id, endpoint, method, status_code, response_time_ms)
		VALUES (?, ?, ?, ?, ?, ?)
	`, keyID, userID, endpoint, method, statusCode, responseTimeMs)
	return err
}

// GetAPIKeys 获取用户的所有API Keys
func (d *Database) GetAPIKeys(userID string) ([]APIKey, error) {
	rows, err := d.db.Query(`
		SELECT id, user_id, key_prefix, name, enabled, rate_limit, usage_count,
		       last_used_at, created_at, expires_at
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		var lastUsedAt sql.NullTime
		var expiresAt sql.NullTime

		err := rows.Scan(
			&key.ID, &key.UserID, &key.KeyPrefix, &key.Name, &key.Enabled,
			&key.RateLimit, &key.UsageCount, &lastUsedAt, &key.CreatedAt, &expiresAt,
		)
		if err != nil {
			return nil, err
		}

		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// RevokeAPIKey 撤销API Key
func (d *Database) RevokeAPIKey(keyID, userID string) error {
	result, err := d.db.Exec(`
		UPDATE api_keys SET enabled = 0 WHERE id = ? AND user_id = ?
	`, keyID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or unauthorized")
	}

	return nil
}

// DeleteAPIKey 删除API Key
func (d *Database) DeleteAPIKey(keyID, userID string) error {
	result, err := d.db.Exec(`
		DELETE FROM api_keys WHERE id = ? AND user_id = ?
	`, keyID, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or unauthorized")
	}

	return nil
}

// GetAPIUsageStats 获取API使用统计
func (d *Database) GetAPIUsageStats(userID string, days int) (map[string]interface{}, error) {
	// 获取总调用次数
	var totalCalls int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM api_usage_logs 
		WHERE user_id = ? AND created_at >= datetime('now', '-' || ? || ' days')
	`, userID, days).Scan(&totalCalls)
	if err != nil {
		return nil, err
	}

	// 获取平均响应时间
	var avgResponseTime sql.NullFloat64
	err = d.db.QueryRow(`
		SELECT AVG(response_time_ms) FROM api_usage_logs 
		WHERE user_id = ? AND created_at >= datetime('now', '-' || ? || ' days')
	`, userID, days).Scan(&avgResponseTime)
	if err != nil {
		return nil, err
	}

	// 获取当前月使用量
	var monthlyUsage int
	err = d.db.QueryRow(`
		SELECT SUM(usage_count) FROM api_keys WHERE user_id = ?
	`, userID).Scan(&monthlyUsage)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_calls":       totalCalls,
		"avg_response_time": avgResponseTime.Float64,
		"monthly_usage":     monthlyUsage,
		"period_days":       days,
	}

	return stats, nil
}
