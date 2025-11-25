package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DecisionLogRecord 决策日志记录（数据库版本，只存储核心数据）
type DecisionLogRecord struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	TraderID            string    `json:"trader_id"`
	CycleNumber         int       `json:"cycle_number"`
	Timestamp           time.Time `json:"timestamp"`
	SystemPrompt        string    `json:"system_prompt"`        // 系统提示词
	InputPrompt         string    `json:"input_prompt"`         // 输入提示
	CoTTrace            string    `json:"cot_trace"`            // AI思维链
	AccountState        string    `json:"account_state"`        // JSON
	Positions           string    `json:"positions"`            // JSON
	Decisions           string    `json:"decisions"`            // JSON
	CandidateCoins      string    `json:"candidate_coins"`      // JSON
	ExecutionLog        string    `json:"execution_log"`        // JSON
	Success             bool      `json:"success"`
	ErrorMessage        string    `json:"error_message"`
	AIRequestDurationMs int64     `json:"ai_request_duration_ms"`
	PromptTokens        int       `json:"prompt_tokens"`
	CompletionTokens    int       `json:"completion_tokens"`
	TotalTokens         int       `json:"total_tokens"`
	CreatedAt           time.Time `json:"created_at"`
}

// SaveDecisionLog 保存决策日志到数据库
func (d *Database) SaveDecisionLog(log *DecisionLogRecord) error {
	// 生成ID
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	
	if d.isPostgreSQL() {
		_, err := d.db.Exec(`
			INSERT INTO decision_logs 
			(id, user_id, trader_id, cycle_number, timestamp, system_prompt, input_prompt, cot_trace,
			 account_state, positions, decisions, candidate_coins, execution_log, success, error_message, ai_request_duration_ms,
			 prompt_tokens, completion_tokens, total_tokens)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		`, log.ID, log.UserID, log.TraderID, log.CycleNumber, log.Timestamp, log.SystemPrompt,
			log.InputPrompt, log.CoTTrace, log.AccountState, log.Positions, log.Decisions, 
			log.CandidateCoins, log.ExecutionLog, log.Success, log.ErrorMessage, log.AIRequestDurationMs,
			log.PromptTokens, log.CompletionTokens, log.TotalTokens)
		
		if err != nil {
			return fmt.Errorf("保存决策日志失败: %w", err)
		}
	} else {
		_, err := d.db.Exec(`
			INSERT INTO decision_logs 
			(id, user_id, trader_id, cycle_number, timestamp, account_state, positions, 
			 decisions, candidate_coins, execution_log, success, error_message, ai_request_duration_ms,
			 prompt_tokens, completion_tokens, total_tokens)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, log.ID, log.UserID, log.TraderID, log.CycleNumber, log.Timestamp, log.AccountState,
			log.Positions, log.Decisions, log.CandidateCoins, log.ExecutionLog, log.Success,
			log.ErrorMessage, log.AIRequestDurationMs, log.PromptTokens, log.CompletionTokens, log.TotalTokens)
		
		if err != nil {
			return fmt.Errorf("保存决策日志失败: %w", err)
		}
	}
	
	return nil
}

// GetDecisionLogs 查询决策日志（分页）
func (d *Database) GetDecisionLogs(userID, traderID string, from, to time.Time, page, limit int) ([]*DecisionLogRecord, int, error) {
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
	
	// 查询总数
	countQuery := d.convertQuery(fmt.Sprintf("SELECT COUNT(*) FROM decision_logs %s", whereClause))
	var total int
	err := d.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询决策日志总数失败: %w", err)
	}
	
	// 查询数据
	offset := (page - 1) * limit
	query := d.convertQuery(fmt.Sprintf(`
		SELECT id, user_id, trader_id, cycle_number, timestamp, account_state, positions, 
		       decisions, candidate_coins, execution_log, success, error_message, 
		       ai_request_duration_ms, created_at
		FROM decision_logs
		%s
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, whereClause))
	
	args = append(args, limit, offset)
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询决策日志失败: %w", err)
	}
	defer rows.Close()
	
	var logs []*DecisionLogRecord
	for rows.Next() {
		var log DecisionLogRecord
		err := rows.Scan(
			&log.ID, &log.UserID, &log.TraderID, &log.CycleNumber, &log.Timestamp,
			&log.AccountState, &log.Positions, &log.Decisions, &log.CandidateCoins,
			&log.ExecutionLog, &log.Success, &log.ErrorMessage, &log.AIRequestDurationMs,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描决策日志失败: %w", err)
		}
		logs = append(logs, &log)
	}
	
	return logs, total, nil
}

// GetLatestDecisionLogs 获取最近N条决策日志
func (d *Database) GetLatestDecisionLogs(userID, traderID string, limit int) ([]*DecisionLogRecord, error) {
	query := d.convertQuery(`
		SELECT id, user_id, trader_id, cycle_number, timestamp, system_prompt, input_prompt, 
		       cot_trace, account_state, positions, decisions, candidate_coins, execution_log, 
		       success, error_message, ai_request_duration_ms, created_at
		FROM decision_logs
		WHERE user_id = ? AND trader_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`)
	
	rows, err := d.db.Query(query, userID, traderID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询最近决策日志失败: %w", err)
	}
	defer rows.Close()
	
	var logs []*DecisionLogRecord
	for rows.Next() {
		var log DecisionLogRecord
		err := rows.Scan(
			&log.ID, &log.UserID, &log.TraderID, &log.CycleNumber, &log.Timestamp,
			&log.SystemPrompt, &log.InputPrompt, &log.CoTTrace,
			&log.AccountState, &log.Positions, &log.Decisions, &log.CandidateCoins,
			&log.ExecutionLog, &log.Success, &log.ErrorMessage, &log.AIRequestDurationMs,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描决策日志失败: %w", err)
		}
		logs = append(logs, &log)
	}
	
	return logs, nil
}

// CleanOldDecisionLogs 清理旧的决策日志
func (d *Database) CleanOldDecisionLogs(days int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	query := d.convertQuery(`DELETE FROM decision_logs WHERE timestamp < ?`)
	result, err := d.db.Exec(query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("清理旧决策日志失败: %w", err)
	}
	
	count, _ := result.RowsAffected()
	return count, nil
}

// GetDecisionLogStatistics 获取决策日志统计信息
func (d *Database) GetDecisionLogStatistics(userID, traderID string, from, to time.Time) (map[string]interface{}, error) {
	// 先准备布尔值参数（PostgreSQL使用true/false）
	successValue := true
	failValue := true
	
	// 布尔值参数必须在最前面，因为它们在SELECT中使用
	args := []interface{}{successValue, failValue}
	
	whereClause := "WHERE user_id = ?"
	args = append(args, userID)
	
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
	
	query := d.convertQuery(fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_cycles,
			SUM(CASE WHEN success = ? THEN 1 ELSE 0 END) as successful_cycles,
			SUM(CASE WHEN success = ? THEN 0 ELSE 1 END) as failed_cycles,
			AVG(ai_request_duration_ms) as avg_ai_duration_ms
		FROM decision_logs
		%s
	`, whereClause))
	
	var totalCycles, successfulCycles, failedCycles int
	var avgAIDuration float64
	
	err := d.db.QueryRow(query, args...).Scan(&totalCycles, &successfulCycles, &failedCycles, &avgAIDuration)
	if err != nil {
		return nil, fmt.Errorf("查询决策日志统计失败: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_cycles":       totalCycles,
		"successful_cycles":  successfulCycles,
		"failed_cycles":      failedCycles,
		"avg_ai_duration_ms": avgAIDuration,
	}
	
	return stats, nil
}

// ConvertToDecisionLogRecord 将logger.DecisionRecord转换为数据库记录
func ConvertToDecisionLogRecord(userID, traderID string, cycleNumber int, record interface{}) (*DecisionLogRecord, error) {
	// 将record序列化为JSON字符串
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("序列化决策记录失败: %w", err)
	}
	
	// 解析为map以提取字段
	var recordMap map[string]interface{}
	if err := json.Unmarshal(recordJSON, &recordMap); err != nil {
		return nil, fmt.Errorf("解析决策记录失败: %w", err)
	}
	
	// 提取字段
	accountStateJSON, _ := json.Marshal(recordMap["account_state"])
	positionsJSON, _ := json.Marshal(recordMap["positions"])
	decisionsJSON, _ := json.Marshal(recordMap["decisions"])
	candidateCoinsJSON, _ := json.Marshal(recordMap["candidate_coins"])
	executionLogJSON, _ := json.Marshal(recordMap["execution_log"])
	
	success, _ := recordMap["success"].(bool)
	errorMessage, _ := recordMap["error_message"].(string)
	aiDuration, _ := recordMap["ai_request_duration_ms"].(float64)
	
	// 解析时间戳
	timestamp := time.Now()
	if ts, ok := recordMap["timestamp"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, ts); err == nil {
			timestamp = parsedTime
		}
	}
	
	return &DecisionLogRecord{
		UserID:              userID,
		TraderID:            traderID,
		CycleNumber:         cycleNumber,
		Timestamp:           timestamp,
		AccountState:        string(accountStateJSON),
		Positions:           string(positionsJSON),
		Decisions:           string(decisionsJSON),
		CandidateCoins:      string(candidateCoinsJSON),
		ExecutionLog:        string(executionLogJSON),
		Success:             success,
		ErrorMessage:        errorMessage,
		AIRequestDurationMs: int64(aiDuration),
	}, nil
}
