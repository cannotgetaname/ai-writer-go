package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BillingRecord 计费记录
type BillingRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	BookName    string    `json:"book_name"`
	ChapterID   int       `json:"chapter_id"`
	TaskType    string    `json:"task_type"`
	Model       string    `json:"model"`
	InputTokens int       `json:"input_tokens"`
	OutputTokens int      `json:"output_tokens"`
	Cost        float64   `json:"cost"`
}

// BillingStats 计费统计
type BillingStats struct {
	TotalTokens   int                 `json:"total_tokens"`
	TotalCost     float64             `json:"total_cost"`
	MonthlyTokens int                 `json:"monthly_tokens"`
	MonthlyCost   float64             `json:"monthly_cost"`
	DailyStats    []DailyStat         `json:"daily_stats"`
	ByModel       map[string]ModelStat `json:"by_model"`
	Records       []BillingRecord     `json:"records"`
}

// DailyStat 每日统计
type DailyStat struct {
	Date    string  `json:"date"`
	Tokens  int     `json:"tokens"`
	Cost    float64 `json:"cost"`
}

// ModelStat 模型统计
type ModelStat struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost"`
}

// BillingStore 计费存储
type BillingStore struct {
	basePath string
	mu       sync.RWMutex
	stats    *BillingStats
}

// NewBillingStore 创建计费存储
func NewBillingStore(basePath string) *BillingStore {
	bs := &BillingStore{
		basePath: basePath,
		stats: &BillingStats{
			ByModel: make(map[string]ModelStat),
			Records: []BillingRecord{},
		},
	}
	bs.load()
	return bs
}

// getBillingFile 获取计费文件路径
func (bs *BillingStore) getBillingFile() string {
	return filepath.Join(bs.basePath, "billing.json")
}

// load 加载计费数据
func (bs *BillingStore) load() {
	data, err := os.ReadFile(bs.getBillingFile())
	if err != nil {
		return
	}

	var stats BillingStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return
	}

	bs.mu.Lock()
	bs.stats = &stats
	if bs.stats.ByModel == nil {
		bs.stats.ByModel = make(map[string]ModelStat)
	}
	bs.mu.Unlock()
}

// save 保存计费数据
func (bs *BillingStore) save() {
	data, err := json.MarshalIndent(bs.stats, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(bs.getBillingFile(), data, 0644)
}

// RecordUsage 记录使用量
func (bs *BillingStore) RecordUsage(bookName, taskType, model string, chapterID, inputTokens, outputTokens int, cost float64) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// 创建记录
	record := BillingRecord{
		Timestamp:    time.Now(),
		BookName:     bookName,
		ChapterID:    chapterID,
		TaskType:     taskType,
		Model:        model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Cost:         cost,
	}

	// 添加记录（保留最近1000条）
	bs.stats.Records = append(bs.stats.Records, record)
	if len(bs.stats.Records) > 1000 {
		bs.stats.Records = bs.stats.Records[len(bs.stats.Records)-1000:]
	}

	// 更新总计
	bs.stats.TotalTokens += inputTokens + outputTokens
	bs.stats.TotalCost += cost

	// 更新月度统计
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	for _, r := range bs.stats.Records {
		if r.Timestamp.After(monthStart) || r.Timestamp.Equal(monthStart) {
			// 已在 MonthlyTokens 中计算
		}
	}
	// 简化：每次添加记录时更新月度数据
	bs.stats.MonthlyTokens += inputTokens + outputTokens
	bs.stats.MonthlyCost += cost

	// 更新模型统计
	modelStat := bs.stats.ByModel[model]
	modelStat.InputTokens += inputTokens
	modelStat.OutputTokens += outputTokens
	modelStat.Cost += cost
	bs.stats.ByModel[model] = modelStat

	// 更新每日统计
	dateStr := now.Format("2006-01-02")
	found := false
	for i, ds := range bs.stats.DailyStats {
		if ds.Date == dateStr {
			bs.stats.DailyStats[i].Tokens += inputTokens + outputTokens
			bs.stats.DailyStats[i].Cost += cost
			found = true
			break
		}
	}
	if !found {
		bs.stats.DailyStats = append(bs.stats.DailyStats, DailyStat{
			Date:   dateStr,
			Tokens: inputTokens + outputTokens,
			Cost:   cost,
		})
	}

	// 保留最近30天的统计
	if len(bs.stats.DailyStats) > 30 {
		bs.stats.DailyStats = bs.stats.DailyStats[len(bs.stats.DailyStats)-30:]
	}

	bs.save()
}

// GetStats 获取统计信息
func (bs *BillingStore) GetStats() *BillingStats {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	// 返回副本
	stats := *bs.stats
	stats.ByModel = make(map[string]ModelStat)
	for k, v := range bs.stats.ByModel {
		stats.ByModel[k] = v
	}

	return &stats
}

// CalculateCost 计算费用
func CalculateCost(model string, inputTokens, outputTokens int, pricing map[string]interface{}) float64 {
	var inputPrice, outputPrice float64

	if p, ok := pricing[model]; ok {
		if pm, ok := p.(map[string]interface{}); ok {
			if ip, ok := pm["input"].(float64); ok {
				inputPrice = ip
			}
			if op, ok := pm["output"].(float64); ok {
				outputPrice = op
			}
		}
	}

	// 价格单位是 $/1K tokens
	cost := (float64(inputTokens)/1000.0)*inputPrice + (float64(outputTokens)/1000.0)*outputPrice
	return cost
}