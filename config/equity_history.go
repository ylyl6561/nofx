package config

import (
	"fmt"
	"time"
)

// EquityHistoryRecord 权益历史记录
type EquityHistoryRecord struct {
	ID                    int64     `json:"id"`
	UserID                string    `json:"user_id"`
	TraderID              string    `json:"trader_id"`
	Timestamp             time.Time `json:"timestamp"`
	TotalEquity           float64   `json:"total_equity"`
	AvailableBalance      float64   `json:"available_balance"`
	TotalUnrealizedProfit float64   `json:"total_unrealized_profit"`
	TotalPnL              float64   `json:"total_pnl"`
	TotalPnLPct           float64   `json:"total_pnl_pct"`
	PositionCount         int       `json:"position_count"`
	MarginUsedPct         float64   `json:"margin_used_pct"`
	CreatedAt             time.Time `json:"created_at"`
}

// SaveEquityHistory 保存权益历史记录
func (d *Database) SaveEquityHistory(record *EquityHistoryRecord) error {
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	
	if d.isPostgreSQL() {
		_, err := d.db.Exec(`
			INSERT INTO equity_history 
			(user_id, trader_id, timestamp, total_equity, available_balance, 
			 total_unrealized_profit, total_pnl, total_pnl_pct, position_count, margin_used_pct)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, record.UserID, record.TraderID, record.Timestamp, record.TotalEquity,
			record.AvailableBalance, record.TotalUnrealizedProfit, record.TotalPnL,
			record.TotalPnLPct, record.PositionCount, record.MarginUsedPct)
		
		if err != nil {
			return fmt.Errorf("保存权益历史失败: %w", err)
		}
	} else {
		_, err := d.db.Exec(`
			INSERT INTO equity_history 
			(user_id, trader_id, timestamp, total_equity, available_balance, 
			 total_unrealized_profit, total_pnl, total_pnl_pct, position_count, margin_used_pct)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, record.UserID, record.TraderID, record.Timestamp, record.TotalEquity,
			record.AvailableBalance, record.TotalUnrealizedProfit, record.TotalPnL,
			record.TotalPnLPct, record.PositionCount, record.MarginUsedPct)
		
		if err != nil {
			return fmt.Errorf("保存权益历史失败: %w", err)
		}
	}
	
	return nil
}

// GetEquityHistory 查询权益历史（分页）
func (d *Database) GetEquityHistory(userID, traderID string, from, to time.Time, interval string, limit int) ([]*EquityHistoryRecord, error) {
	// 构建查询条件
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}
	
	if traderID != "" {
		whereClause += " AND trader_id = ?"
		args = append(args, traderID)
	}
	
	if !from.IsZero() {
		whereClause += " AND timestamp >= ?"
		args = append(args, from)
	}
	
	if !to.IsZero() {
		whereClause += " AND timestamp <= ?"
		args = append(args, to)
	}
	
	// 根据interval进行采样（简化版：直接查询，前端处理采样）
	query := d.convertQuery(fmt.Sprintf(`
		SELECT id, user_id, trader_id, timestamp, total_equity, available_balance, 
		       total_unrealized_profit, total_pnl, total_pnl_pct, position_count, 
		       margin_used_pct, created_at
		FROM equity_history
		%s
		ORDER BY timestamp DESC
		LIMIT ?
	`, whereClause))
	
	args = append(args, limit)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询权益历史失败: %w", err)
	}
	defer rows.Close()
	
	var records []*EquityHistoryRecord
	for rows.Next() {
		var record EquityHistoryRecord
		err := rows.Scan(
			&record.ID, &record.UserID, &record.TraderID, &record.Timestamp,
			&record.TotalEquity, &record.AvailableBalance, &record.TotalUnrealizedProfit,
			&record.TotalPnL, &record.TotalPnLPct, &record.PositionCount,
			&record.MarginUsedPct, &record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描权益历史失败: %w", err)
		}
		records = append(records, &record)
	}
	
	// 反转数组，让时间从旧到新排列（用于图表显示）
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
	
	return records, nil
}

// GetLatestEquityHistory 获取最新的权益历史记录
func (d *Database) GetLatestEquityHistory(userID, traderID string) (*EquityHistoryRecord, error) {
	query := d.convertQuery(`
		SELECT id, user_id, trader_id, timestamp, total_equity, available_balance, 
		       total_unrealized_profit, total_pnl, total_pnl_pct, position_count, 
		       margin_used_pct, created_at
		FROM equity_history
		WHERE user_id = ? AND trader_id = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`)
	
	var record EquityHistoryRecord
	err := d.db.QueryRow(query, userID, traderID).Scan(
		&record.ID, &record.UserID, &record.TraderID, &record.Timestamp,
		&record.TotalEquity, &record.AvailableBalance, &record.TotalUnrealizedProfit,
		&record.TotalPnL, &record.TotalPnLPct, &record.PositionCount,
		&record.MarginUsedPct, &record.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("查询最新权益历史失败: %w", err)
	}
	
	return &record, nil
}

// CleanOldEquityHistory 清理旧的权益历史记录
func (d *Database) CleanOldEquityHistory(days int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	query := d.convertQuery(`DELETE FROM equity_history WHERE timestamp < ?`)
	result, err := d.db.Exec(query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("清理旧权益历史失败: %w", err)
	}
	
	count, _ := result.RowsAffected()
	return count, nil
}

// GetEquityHistoryBatch 批量查询多个交易员的权益历史
func (d *Database) GetEquityHistoryBatch(userID string, traderIDs []string, from, to time.Time, limit int) (map[string][]*EquityHistoryRecord, error) {
	result := make(map[string][]*EquityHistoryRecord)
	
	for _, traderID := range traderIDs {
		records, err := d.GetEquityHistory(userID, traderID, from, to, "", limit)
		if err != nil {
			// 跳过错误，继续查询其他交易员
			continue
		}
		result[traderID] = records
	}
	
	return result, nil
}
