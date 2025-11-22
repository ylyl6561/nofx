package cleaner

import (
	"log"
	"nofx/config"
	"strconv"
	"time"
)

// DataCleaner 数据清理器
type DataCleaner struct {
	database *config.Database
	stopCh   chan struct{}
}

// NewDataCleaner 创建数据清理器
func NewDataCleaner(database *config.Database) *DataCleaner {
	return &DataCleaner{
		database: database,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动定时清理任务
func (dc *DataCleaner) Start() {
	log.Println("🧹 启动数据清理任务...")
	
	// 启动时立即执行一次清理
	go dc.cleanupOldData()
	
	// 每天凌晨3点执行清理
	ticker := time.NewTicker(24 * time.Hour)
	
	// 计算到下一个凌晨3点的时间
	now := time.Now()
	next3AM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if now.After(next3AM) {
		// 如果现在已经过了今天的3点，设置为明天的3点
		next3AM = next3AM.Add(24 * time.Hour)
	}
	
	// 等待到第一个3点
	initialDelay := time.Until(next3AM)
	log.Printf("🧹 下次数据清理时间: %s (距离现在 %.1f 小时)", next3AM.Format("2006-01-02 15:04:05"), initialDelay.Hours())
	
	go func() {
		// 等待到第一个3点
		select {
		case <-time.After(initialDelay):
			dc.cleanupOldData()
		case <-dc.stopCh:
			return
		}
		
		// 之后每24小时执行一次
		for {
			select {
			case <-ticker.C:
				dc.cleanupOldData()
			case <-dc.stopCh:
				ticker.Stop()
				log.Println("🧹 数据清理任务已停止")
				return
			}
		}
	}()
}

// Stop 停止清理任务
func (dc *DataCleaner) Stop() {
	close(dc.stopCh)
}

// cleanupOldData 清理旧数据
func (dc *DataCleaner) cleanupOldData() {
	log.Println("🧹 开始清理旧数据...")
	
	// 获取保留天数配置
	decisionLogsDaysStr, _ := dc.database.GetSystemConfig("decision_logs_retention_days")
	equityHistoryDaysStr, _ := dc.database.GetSystemConfig("equity_history_retention_days")
	
	decisionLogsDays := 10 // 默认10天
	if val, err := strconv.Atoi(decisionLogsDaysStr); err == nil && val > 0 {
		decisionLogsDays = val
	}
	
	equityHistoryDays := 90 // 默认90天
	if val, err := strconv.Atoi(equityHistoryDaysStr); err == nil && val > 0 {
		equityHistoryDays = val
	}
	
	log.Printf("🧹 配置: 决策日志保留 %d 天, 权益历史保留 %d 天", decisionLogsDays, equityHistoryDays)
	
	// 清理决策日志
	if count, err := dc.database.CleanOldDecisionLogs(decisionLogsDays); err == nil {
		if count > 0 {
			log.Printf("🗑️  清理了 %d 条旧决策日志 (%d天前)", count, decisionLogsDays)
		} else {
			log.Printf("✓ 无需清理决策日志")
		}
	} else {
		log.Printf("❌ 清理决策日志失败: %v", err)
	}
	
	// 清理权益历史
	if count, err := dc.database.CleanOldEquityHistory(equityHistoryDays); err == nil {
		if count > 0 {
			log.Printf("🗑️  清理了 %d 条旧权益历史 (%d天前)", count, equityHistoryDays)
		} else {
			log.Printf("✓ 无需清理权益历史")
		}
	} else {
		log.Printf("❌ 清理权益历史失败: %v", err)
	}
	
	log.Println("🧹 数据清理完成")
}

// CleanupNow 立即执行清理（用于手动触发）
func (dc *DataCleaner) CleanupNow() {
	dc.cleanupOldData()
}
