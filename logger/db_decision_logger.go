package logger

import (
	"encoding/json"
	"fmt"
	"math"
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
	
	analysis := &PerformanceAnalysis{
		RecentTrades: []TradeOutcome{},
		SymbolStats:  make(map[string]*SymbolPerformance),
	}
	
	// 追踪持仓状态：symbol_side -> {side, openPrice, openTime, quantity, leverage}
	openPositions := make(map[string]map[string]interface{})
	
	// 为了避免开仓记录在窗口外导致匹配失败，需要先从所有历史记录中找出未平仓的持仓
	// 获取更多历史记录来构建完整的持仓状态（使用更大的窗口）
	allRecords, err := l.GetLatestRecords(lookbackCycles * 3) // 扩大3倍窗口
	if err == nil && len(allRecords) > len(records) {
		// 先从扩大的窗口中收集所有开仓记录
		for _, record := range allRecords {
			for _, action := range record.Decisions {
				if !action.Success {
					continue
				}
				
				symbol := action.Symbol
				side := ""
				if action.Action == "open_long" || action.Action == "close_long" || action.Action == "partial_close" || action.Action == "auto_close_long" {
					side = "long"
				} else if action.Action == "open_short" || action.Action == "close_short" || action.Action == "auto_close_short" {
					side = "short"
				}
				
				// partial_close 需要根據持倉判斷方向
				if action.Action == "partial_close" && side == "" {
					for key, pos := range openPositions {
						if posSymbol, _ := pos["side"].(string); key == symbol+"_"+posSymbol {
							side = posSymbol
							break
						}
					}
				}
				
				posKey := symbol + "_" + side
				
				switch action.Action {
				case "open_long", "open_short":
					// 记录开仓
					openPositions[posKey] = map[string]interface{}{
						"side":      side,
						"openPrice": action.Price,
						"openTime":  action.Timestamp,
						"quantity":  action.Quantity,
						"leverage":  action.Leverage,
					}
				case "close_long", "close_short", "auto_close_long", "auto_close_short":
					// 移除已平仓记录
					delete(openPositions, posKey)
					// partial_close 不處理，保留持倉記錄
				}
			}
		}
	}
	
	// 遍历分析窗口内的记录，生成交易结果
	for _, record := range records {
		for _, action := range record.Decisions {
			if !action.Success {
				continue
			}
			
			symbol := action.Symbol
			side := ""
			if action.Action == "open_long" || action.Action == "close_long" || action.Action == "partial_close" || action.Action == "auto_close_long" {
				side = "long"
			} else if action.Action == "open_short" || action.Action == "close_short" || action.Action == "auto_close_short" {
				side = "short"
			}
			
			// partial_close 需要根據持倉判斷方向
			if action.Action == "partial_close" {
				// 從 openPositions 中查找持倉方向
				for key, pos := range openPositions {
					if posSymbol, _ := pos["side"].(string); key == symbol+"_"+posSymbol {
						side = posSymbol
						break
					}
				}
			}
			
			posKey := symbol + "_" + side // 使用symbol_side作为key，区分多空持仓
			
			switch action.Action {
			case "open_long", "open_short":
				// 更新开仓记录（可能已经在预填充时记录过了）
				openPositions[posKey] = map[string]interface{}{
					"side":               side,
					"openPrice":          action.Price,
					"openTime":           action.Timestamp,
					"quantity":           action.Quantity,
					"leverage":           action.Leverage,
					"remainingQuantity":  action.Quantity, // 追蹤剩餘數量
					"accumulatedPnL":     0.0,             // 累積部分平倉盈虧
					"partialCloseCount":  0,               // 部分平倉次數
					"partialCloseVolume": 0.0,             // 部分平倉總量
				}
				
			case "close_long", "close_short", "partial_close", "auto_close_long", "auto_close_short":
				// 查找对应的开仓记录（可能来自预填充或当前窗口）
				if openPos, exists := openPositions[posKey]; exists {
					openPrice := openPos["openPrice"].(float64)
					openTime := openPos["openTime"].(time.Time)
					side := openPos["side"].(string)
					quantity := openPos["quantity"].(float64)
					leverage := openPos["leverage"].(int)
					
					// 取得追蹤字段（若不存在則初始化）
					remainingQty, _ := openPos["remainingQuantity"].(float64)
					if remainingQty == 0 {
						remainingQty = quantity // 兼容舊數據（沒有 remainingQuantity 字段）
					}
					accumulatedPnL, _ := openPos["accumulatedPnL"].(float64)
					partialCloseCount, _ := openPos["partialCloseCount"].(int)
					partialCloseVolume, _ := openPos["partialCloseVolume"].(float64)
					
					// 对于 partial_close，使用实际平仓数量；否则使用剩余仓位数量
					actualQuantity := remainingQty
					if action.Action == "partial_close" {
						actualQuantity = action.Quantity
					}
					
					// 计算本次平仓的盈亏（USDT）
					var pnl float64
					if side == "long" {
						pnl = actualQuantity * (action.Price - openPrice)
					} else {
						pnl = actualQuantity * (openPrice - action.Price)
					}
					
					// 處理 partial_close 聚合邏輯
					if action.Action == "partial_close" {
						// 累積盈虧和數量
						accumulatedPnL += pnl
						remainingQty -= actualQuantity
						partialCloseCount++
						partialCloseVolume += actualQuantity
						
						// 更新 openPositions（保留持倉記錄，但更新追蹤數據）
						openPos["remainingQuantity"] = remainingQty
						openPos["accumulatedPnL"] = accumulatedPnL
						openPos["partialCloseCount"] = partialCloseCount
						openPos["partialCloseVolume"] = partialCloseVolume
						
						// 判斷是否已完全平倉
						if remainingQty <= 0.0001 { // 使用小閾值避免浮點誤差
							// 完全平倉：記錄為一筆完整交易
							positionValue := quantity * openPrice
							marginUsed := positionValue / float64(leverage)
							pnlPct := 0.0
							if marginUsed > 0 {
								pnlPct = (accumulatedPnL / marginUsed) * 100
							}
							
							outcome := TradeOutcome{
								Symbol:        symbol,
								Side:          side,
								Quantity:      quantity, // 使用原始總量
								Leverage:      leverage,
								OpenPrice:     openPrice,
								ClosePrice:    action.Price, // 最後一次平倉價格
								PositionValue: positionValue,
								MarginUsed:    marginUsed,
								PnL:           accumulatedPnL, // 使用累積盈虧
								PnLPct:        pnlPct,
								Duration:      action.Timestamp.Sub(openTime).String(),
								OpenTime:      openTime,
								CloseTime:     action.Timestamp,
							}
							
							analysis.RecentTrades = append(analysis.RecentTrades, outcome)
							analysis.TotalTrades++ // 只在完全平倉時計數
							
							// 分类交易
							if accumulatedPnL > 0 {
								analysis.WinningTrades++
								analysis.AvgWin += accumulatedPnL
							} else if accumulatedPnL < 0 {
								analysis.LosingTrades++
								analysis.AvgLoss += accumulatedPnL
							}
							
							// 更新币种统计
							if _, exists := analysis.SymbolStats[symbol]; !exists {
								analysis.SymbolStats[symbol] = &SymbolPerformance{
									Symbol: symbol,
								}
							}
							stats := analysis.SymbolStats[symbol]
							stats.TotalTrades++
							stats.TotalPnL += accumulatedPnL
							if accumulatedPnL > 0 {
								stats.WinningTrades++
							} else if accumulatedPnL < 0 {
								stats.LosingTrades++
							}
							
							// 刪除持倉記錄
							delete(openPositions, posKey)
						}
						// 否則不做任何操作（等待後續 partial_close 或 full close）
						
					} else {
						// 完全平倉（close_long/close_short/auto_close）
						// 如果之前有部分平倉，需要加上累積的 PnL
						totalPnL := accumulatedPnL + pnl
						
						positionValue := quantity * openPrice
						marginUsed := positionValue / float64(leverage)
						pnlPct := 0.0
						if marginUsed > 0 {
							pnlPct = (totalPnL / marginUsed) * 100
						}
						
						outcome := TradeOutcome{
							Symbol:        symbol,
							Side:          side,
							Quantity:      quantity, // 使用原始總量
							Leverage:      leverage,
							OpenPrice:     openPrice,
							ClosePrice:    action.Price,
							PositionValue: positionValue,
							MarginUsed:    marginUsed,
							PnL:           totalPnL, // 包含之前部分平倉的 PnL
							PnLPct:        pnlPct,
							Duration:      action.Timestamp.Sub(openTime).String(),
							OpenTime:      openTime,
							CloseTime:     action.Timestamp,
						}
						
						analysis.RecentTrades = append(analysis.RecentTrades, outcome)
						analysis.TotalTrades++
						
						// 分类交易
						if totalPnL > 0 {
							analysis.WinningTrades++
							analysis.AvgWin += totalPnL
						} else if totalPnL < 0 {
							analysis.LosingTrades++
							analysis.AvgLoss += totalPnL
						}
						
						// 更新币种统计
						if _, exists := analysis.SymbolStats[symbol]; !exists {
							analysis.SymbolStats[symbol] = &SymbolPerformance{
								Symbol: symbol,
							}
						}
						stats := analysis.SymbolStats[symbol]
						stats.TotalTrades++
						stats.TotalPnL += totalPnL
						if totalPnL > 0 {
							stats.WinningTrades++
						} else if totalPnL < 0 {
							stats.LosingTrades++
						}
						
						// 刪除持倉記錄
						delete(openPositions, posKey)
					}
				}
			}
		}
	}
	
	// 计算统计指标
	if analysis.TotalTrades > 0 {
		analysis.WinRate = (float64(analysis.WinningTrades) / float64(analysis.TotalTrades)) * 100
		
		// 计算总盈利和总亏损
		totalWinAmount := analysis.AvgWin   // 当前是累加的总和
		totalLossAmount := analysis.AvgLoss // 当前是累加的总和（负数）
		
		if analysis.WinningTrades > 0 {
			analysis.AvgWin /= float64(analysis.WinningTrades)
		}
		if analysis.LosingTrades > 0 {
			analysis.AvgLoss /= float64(analysis.LosingTrades)
		}
		
		// Profit Factor = 总盈利 / 总亏损（绝对值）
		// 注意：totalLossAmount 是负数，所以取负号得到绝对值
		if totalLossAmount != 0 {
			analysis.ProfitFactor = totalWinAmount / (-totalLossAmount)
		} else if totalWinAmount > 0 {
			// 只有盈利没有亏损的情况，设置为一个很大的值表示完美策略
			analysis.ProfitFactor = 999.0
		}
	}
	
	// 计算各币种胜率和平均盈亏
	bestPnL := -999999.0
	worstPnL := 999999.0
	for symbol, stats := range analysis.SymbolStats {
		if stats.TotalTrades > 0 {
			stats.WinRate = (float64(stats.WinningTrades) / float64(stats.TotalTrades)) * 100
			stats.AvgPnL = stats.TotalPnL / float64(stats.TotalTrades)
			
			if stats.TotalPnL > bestPnL {
				bestPnL = stats.TotalPnL
				analysis.BestSymbol = symbol
			}
			if stats.TotalPnL < worstPnL {
				worstPnL = stats.TotalPnL
				analysis.WorstSymbol = symbol
			}
		}
	}
	
	// 只保留最近的交易（倒序：最新的在前）
	if len(analysis.RecentTrades) > 10 {
		// 反转数组，让最新的在前
		for i, j := 0, len(analysis.RecentTrades)-1; i < j; i, j = i+1, j-1 {
			analysis.RecentTrades[i], analysis.RecentTrades[j] = analysis.RecentTrades[j], analysis.RecentTrades[i]
		}
		analysis.RecentTrades = analysis.RecentTrades[:10]
	} else if len(analysis.RecentTrades) > 0 {
		// 反转数组
		for i, j := 0, len(analysis.RecentTrades)-1; i < j; i, j = i+1, j-1 {
			analysis.RecentTrades[i], analysis.RecentTrades[j] = analysis.RecentTrades[j], analysis.RecentTrades[i]
		}
	}
	
	// 计算夏普比率（需要至少2个数据点）
	analysis.SharpeRatio = l.calculateSharpeRatio(records)
	
	return analysis, nil
}

// calculateSharpeRatio 计算夏普比率
// 基于账户净值的变化计算风险调整后收益
func (l *DBDecisionLogger) calculateSharpeRatio(records []*DecisionRecord) float64 {
	if len(records) < 2 {
		return 0.0
	}
	
	// 提取每个周期的账户净值
	// 注意：TotalBalance字段实际存储的是TotalEquity（账户总净值）
	var equities []float64
	for _, record := range records {
		// 直接使用TotalBalance，因为它已经是完整的账户净值
		equity := record.AccountState.TotalBalance
		if equity > 0 {
			equities = append(equities, equity)
		}
	}
	
	if len(equities) < 2 {
		return 0.0
	}
	
	// 计算周期收益率（period returns）
	var returns []float64
	for i := 1; i < len(equities); i++ {
		if equities[i-1] > 0 {
			periodReturn := (equities[i] - equities[i-1]) / equities[i-1]
			returns = append(returns, periodReturn)
		}
	}
	
	if len(returns) == 0 {
		return 0.0
	}
	
	// 计算平均收益率
	sumReturns := 0.0
	for _, r := range returns {
		sumReturns += r
	}
	meanReturn := sumReturns / float64(len(returns))
	
	// 计算收益率标准差
	sumSquaredDiff := 0.0
	for _, r := range returns {
		diff := r - meanReturn
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(returns))
	stdDev := math.Sqrt(variance)
	
	// 避免除以零
	if stdDev == 0 {
		if meanReturn > 0 {
			return 999.0 // 无波动的正收益
		} else if meanReturn < 0 {
			return -999.0 // 无波动的负收益
		}
		return 0.0
	}
	
	// 计算夏普比率（假设无风险利率为0）
	// 注：直接返回周期级别的夏普比率（非年化），正常范围 -2 到 +2
	sharpeRatio := meanReturn / stdDev
	return sharpeRatio
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
