package logger

import (
	"encoding/json"
	"fmt"
	"nofx/config"
	"time"
)

// DBDecisionLogger 数据库决策日志记录器
type DBDecisionLogger struct {
	database    *config.Database
	userID      string
	traderID    string
	cycleNumber int
}

// NewDBDecisionLogger 创建数据库决策日志记录器
func NewDBDecisionLogger(database *config.Database, userID, traderID string) *DBDecisionLogger {
	return &DBDecisionLogger{
		database:    database,
		userID:      userID,
		traderID:    traderID,
		cycleNumber: 0,
	}
}

// LogDecision 记录决策到数据库
func (l *DBDecisionLogger) LogDecision(record *DecisionRecord) error {
	l.cycleNumber++
	record.CycleNumber = l.cycleNumber
	record.Timestamp = time.Now()
	
	// 序列化各个字段为JSON
	accountStateJSON, err := json.Marshal(record.AccountState)
	if err != nil {
		return fmt.Errorf("序列化账户状态失败: %w", err)
	}
	
	positionsJSON, err := json.Marshal(record.Positions)
	if err != nil {
		return fmt.Errorf("序列化持仓列表失败: %w", err)
	}
	
	decisionsJSON, err := json.Marshal(record.Decisions)
	if err != nil {
		return fmt.Errorf("序列化决策列表失败: %w", err)
	}
	
	candidateCoinsJSON, err := json.Marshal(record.CandidateCoins)
	if err != nil {
		return fmt.Errorf("序列化候选币种失败: %w", err)
	}
	
	executionLogJSON, err := json.Marshal(record.ExecutionLog)
	if err != nil {
		return fmt.Errorf("序列化执行日志失败: %w", err)
	}
	
	// 创建数据库记录
	dbRecord := &config.DecisionLogRecord{
		UserID:              l.userID,
		TraderID:            l.traderID,
		CycleNumber:         record.CycleNumber,
		Timestamp:           record.Timestamp,
		SystemPrompt:        record.SystemPrompt,
		InputPrompt:         record.InputPrompt,
		CoTTrace:            record.CoTTrace,
		AccountState:        string(accountStateJSON),
		Positions:           string(positionsJSON),
		Decisions:           string(decisionsJSON),
		CandidateCoins:      string(candidateCoinsJSON),
		ExecutionLog:        string(executionLogJSON),
		Success:             record.Success,
		ErrorMessage:        record.ErrorMessage,
		AIRequestDurationMs: record.AIRequestDurationMs,
	}
	
	// 保存到数据库
	if err := l.database.SaveDecisionLog(dbRecord); err != nil {
		return fmt.Errorf("保存决策日志到数据库失败: %w", err)
	}
	
	fmt.Printf("📝 决策记录已保存到数据库: trader=%s cycle=%d\n", l.traderID, record.CycleNumber)
	return nil
}

// GetLatestRecords 获取最近N条记录（按时间正序：从旧到新）
func (l *DBDecisionLogger) GetLatestRecords(n int) ([]*DecisionRecord, error) {
	dbRecords, err := l.database.GetLatestDecisionLogs(l.userID, l.traderID, n)
	if err != nil {
		return nil, err
	}
	
	// 转换为DecisionRecord
	var records []*DecisionRecord
	for _, dbRecord := range dbRecords {
		record, err := l.convertFromDBRecord(dbRecord)
		if err != nil {
			// 跳过转换失败的记录
			continue
		}
		records = append(records, record)
	}
	
	// 反转数组，让时间从旧到新排列
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
	
	return records, nil
}

// GetRecordByDate 获取指定日期的所有记录
func (l *DBDecisionLogger) GetRecordByDate(date time.Time) ([]*DecisionRecord, error) {
	// 构建日期范围
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.Add(24 * time.Hour)
	
	dbRecords, _, err := l.database.GetDecisionLogs(l.userID, l.traderID, from, to, 1, 1000)
	if err != nil {
		return nil, err
	}
	
	// 转换为DecisionRecord
	var records []*DecisionRecord
	for _, dbRecord := range dbRecords {
		record, err := l.convertFromDBRecord(dbRecord)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	
	return records, nil
}

// CleanOldRecords 清理N天前的旧记录
func (l *DBDecisionLogger) CleanOldRecords(days int) error {
	count, err := l.database.CleanOldDecisionLogs(days)
	if err != nil {
		return err
	}
	
	if count > 0 {
		fmt.Printf("🗑️ 已清理 %d 条旧决策记录（%d天前）\n", count, days)
	}
	
	return nil
}

// GetStatistics 获取统计信息
func (l *DBDecisionLogger) GetStatistics() (*Statistics, error) {
	stats, err := l.database.GetDecisionLogStatistics(l.userID, l.traderID, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	
	// 转换为Statistics结构
	result := &Statistics{
		TotalCycles:      stats["total_cycles"].(int),
		SuccessfulCycles: stats["successful_cycles"].(int),
		FailedCycles:     stats["failed_cycles"].(int),
	}
	
	// 注意：数据库版本不统计开仓/平仓次数，需要从decisions字段解析
	// 这里简化处理，返回基础统计
	
	return result, nil
}

// AnalyzePerformance 分析最近N个周期的交易表现
func (l *DBDecisionLogger) AnalyzePerformance(lookbackCycles int) (*PerformanceAnalysis, error) {
	// 获取最近的记录
	records, err := l.GetLatestRecords(lookbackCycles)
	if err != nil {
		return nil, err
	}
	
	if len(records) == 0 {
		return &PerformanceAnalysis{
			RecentTrades: []TradeOutcome{},
			SymbolStats:  make(map[string]*SymbolPerformance),
		}, nil
	}
	
	// 复用原有的分析逻辑
	// 注意：这里需要实现完整的交易分析逻辑，与原logger/decision_logger.go中的AnalyzePerformance类似
	// 为了简化，这里返回基础结构
	
	analysis := &PerformanceAnalysis{
		RecentTrades: []TradeOutcome{},
		SymbolStats:  make(map[string]*SymbolPerformance),
	}
	
	// TODO: 实现完整的交易分析逻辑
	// 可以参考原有的AnalyzePerformance方法
	
	return analysis, nil
}

// convertFromDBRecord 将数据库记录转换为DecisionRecord
func (l *DBDecisionLogger) convertFromDBRecord(dbRecord *config.DecisionLogRecord) (*DecisionRecord, error) {
	record := &DecisionRecord{
		Timestamp:           dbRecord.Timestamp,
		CycleNumber:         dbRecord.CycleNumber,
		Success:             dbRecord.Success,
		ErrorMessage:        dbRecord.ErrorMessage,
		AIRequestDurationMs: dbRecord.AIRequestDurationMs,
		SystemPrompt:        dbRecord.SystemPrompt,
		InputPrompt:         dbRecord.InputPrompt,
		CoTTrace:            dbRecord.CoTTrace,
	}
	
	// 解析JSON字段
	if err := json.Unmarshal([]byte(dbRecord.AccountState), &record.AccountState); err != nil {
		return nil, fmt.Errorf("解析账户状态失败: %w", err)
	}
	
	if err := json.Unmarshal([]byte(dbRecord.Positions), &record.Positions); err != nil {
		return nil, fmt.Errorf("解析持仓列表失败: %w", err)
	}
	
	if err := json.Unmarshal([]byte(dbRecord.Decisions), &record.Decisions); err != nil {
		return nil, fmt.Errorf("解析决策列表失败: %w", err)
	}
	
	if err := json.Unmarshal([]byte(dbRecord.CandidateCoins), &record.CandidateCoins); err != nil {
		return nil, fmt.Errorf("解析候选币种失败: %w", err)
	}
	
	if err := json.Unmarshal([]byte(dbRecord.ExecutionLog), &record.ExecutionLog); err != nil {
		return nil, fmt.Errorf("解析执行日志失败: %w", err)
	}
	
	return record, nil
}
