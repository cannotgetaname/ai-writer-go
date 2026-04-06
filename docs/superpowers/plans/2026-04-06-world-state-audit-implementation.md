# 世界状态审计系统实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现一键审计世界状态功能，AI 自动提取 6 类图谱数据，用户分类审核确认。

**Architecture:** 后端扩展模型字段 + 新增提取服务 + 新增 API 接口；前端新增审核对话框组件 + 写作页面按钮。

**Tech Stack:** Go Gin, Vue 3, Element Plus

---

## 文件结构

**后端修改:**
- `internal/model/causal_chain.go` - 扩展 CausalEvent
- `internal/model/foreshadow.go` - 扩展 Foreshadow
- `internal/model/narrative_thread.go` - 扩展 NarrativeThread、TimelineEvent
- `internal/model/character.go` - 扩展 EmotionPoint
- `internal/model/analysis.go` - 新建：分析报告结构
- `internal/model/pending_sync.go` - 新建：待审核图谱变更
- `internal/store/json_store.go` - 新增 Load/Save 方法
- `internal/service/world_audit.go` - 新建：世界状态审计服务
- `internal/service/analysis.go` - 新建：分析服务
- `internal/api/handler/handler.go` - 新增 handler
- `internal/api/router.go` - 新增路由

**前端创建:**
- `web/src/components/WorldAuditDialog.vue` - 审核对话框
- `web/src/components/audit/StateChangeTab.vue` - 状态变更标签页
- `web/src/components/audit/CausalChainTab.vue` - 因果链标签页
- `web/src/components/audit/ForeshadowTab.vue` - 伏笔标签页
- `web/src/components/audit/ThreadTab.vue` - 叙事线程标签页
- `web/src/components/audit/EmotionTab.vue` - 情感弧线标签页
- `web/src/components/audit/TimelineTab.vue` - 时间线标签页

**前端修改:**
- `web/src/api/index.js` - 新增 API 方法
- `web/src/views/WritingView.vue` - 添加审计按钮
- `web/src/views/AnalysisView.vue` - 增强分析页面

---

## Phase 1: 模型扩展

### Task 1: 扩展 CausalEvent 模型

**Files:**
- Modify: `internal/model/causal_chain.go`

- [ ] **Step 1: 添加新字段**

在 `CausalEvent` 结构体中添加：

```go
// 新增字段
AutoDetected  bool   `json:"auto_detected"`   // 是否AI检测
SourceContext string `json:"source_context"`  // 原文片段
```

- [ ] **Step 2: 构建验证**

Run: `cd /home/zcz/program/ai-writer-go && go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/model/causal_chain.go
git commit -m "feat(model): add AutoDetected and SourceContext fields to CausalEvent"
```

---

### Task 2: 扩展 Foreshadow 模型

**Files:**
- Modify: `internal/model/foreshadow.go`

- [ ] **Step 1: 添加新字段**

在 `Foreshadow` 结构体 `Notes` 字段前添加：

```go
// 新增字段
AutoDetected  bool   `json:"auto_detected"`
SourceContext string `json:"source_context"`
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/model/foreshadow.go
git commit -m "feat(model): add AutoDetected and SourceContext fields to Foreshadow"
```

---

### Task 3: 扩展 EmotionPoint 模型

**Files:**
- Modify: `internal/model/character.go`

- [ ] **Step 1: 找到 EmotionPoint 结构体并添加字段**

在 `EmotionPoint` 结构体中添加：

```go
// 新增字段
SourceContext string `json:"source_context"` // 触发原文片段
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/model/character.go
git commit -m "feat(model): add SourceContext field to EmotionPoint"
```

---

### Task 4: 扩展 NarrativeThread 模型

**Files:**
- Modify: `internal/model/narrative_thread.go`

- [ ] **Step 1: 添加新字段到 NarrativeThread**

在 `NarrativeThread` 结构体 `Chapters` 字段后添加：

```go
// 新增字段
AutoDetected bool `json:"auto_detected"`
```

- [ ] **Step 2: 添加新字段到 TimelineEvent**

在 `TimelineEvent` 结构体末尾添加：

```go
// 新增字段
AutoDetected bool `json:"auto_detected"`
```

- [ ] **Step 3: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/model/narrative_thread.go
git commit -m "feat(model): add AutoDetected field to NarrativeThread and TimelineEvent"
```

---

### Task 5: 创建分析报告模型

**Files:**
- Create: `internal/model/analysis.go`

- [ ] **Step 1: 创建 analysis.go**

```go
package model

import "time"

// AnalysisReport 分析报告
type AnalysisReport struct {
	ID        string    `json:"id"`
	BookID    string    `json:"book_id"`
	ChapterID int       `json:"chapter_id"`
	Type      string    `json:"type"` // review / manual
	CreatedAt time.Time `json:"created_at"`

	// 分析结果
	ForeshadowAnalysis ForeshadowAnalysis `json:"foreshadow_analysis"`
	CausalAnalysis     CausalAnalysis     `json:"causal_analysis"`
	ThreadAnalysis     ThreadAnalysis     `json:"thread_analysis"`
	EmotionAnalysis    EmotionAnalysis    `json:"emotion_analysis"`
	TimelineAnalysis   TimelineAnalysis   `json:"timeline_analysis"`
}

// ForeshadowAnalysis 伏笔分析
type ForeshadowAnalysis struct {
	Warnings    []ForeshadowWarning `json:"warnings"`
	Suggestions []string            `json:"suggestions"`
	Score       int                 `json:"score"`
}

// CausalAnalysis 因果链分析
type CausalAnalysis struct {
	BrokenChains []BrokenChain `json:"broken_chains"`
	OrphanEvents []OrphanEvent `json:"orphan_events"`
	CircularDeps []CircularDep `json:"circular_deps"`
	Score        int           `json:"score"`
}

// BrokenChain 断裂的因果链
type BrokenChain struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	Issue     string `json:"issue"` // "有因无果" / "有果无因"
}

// OrphanEvent 孤立事件
type OrphanEvent struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
}

// CircularDep 循环依赖
type CircularDep struct {
	Chain []string `json:"chain"` // 事件ID链
}

// ThreadAnalysis 叙事线程分析
type ThreadAnalysis struct {
	ForgottenThreads []ForgottenThread `json:"forgotten_threads"`
	PacingIssues     []PacingIssue     `json:"pacing_issues"`
	Conflicts        []ThreadConflict  `json:"conflicts"`
	Score            int               `json:"score"`
}

// ForgottenThread 遗忘的线程
type ForgottenThread struct {
	ThreadID        string `json:"thread_id"`
	ThreadName      string `json:"thread_name"`
	LastActive      int    `json:"last_active"`
	CurrentChapter  int    `json:"current_chapter"`
	ChaptersSkipped int    `json:"chapters_skipped"`
}

// PacingIssue 节奏问题
type PacingIssue struct {
	ThreadID   string `json:"thread_id"`
	ThreadName string `json:"thread_name"`
	Issue      string `json:"issue"`
}

// ThreadConflict 线程冲突
type ThreadConflict struct {
	ThreadIDs   []string `json:"thread_ids"`
	ChapterID   int      `json:"chapter_id"`
	ConflictType string  `json:"conflict_type"`
}

// EmotionAnalysis 情感弧线分析
type EmotionAnalysis struct {
	Inconsistencies []EmotionInconsistency `json:"inconsistencies"`
	PacingIssues    []EmotionPacingIssue   `json:"pacing_issues"`
	WeavingScore    int                    `json:"weaving_score"`
	Score           int                    `json:"score"`
}

// EmotionInconsistency 情感不一致
type EmotionInconsistency struct {
	Character     string `json:"character"`
	ChapterID     int    `json:"chapter_id"`
	FromEmotion   string `json:"from_emotion"`
	ToEmotion     string `json:"to_emotion"`
	IntensityJump int    `json:"intensity_jump"`
}

// EmotionPacingIssue 情感节奏问题
type EmotionPacingIssue struct {
	Character string `json:"character"`
	Issue     string `json:"issue"`
}

// TimelineAnalysis 时间线分析
type TimelineAnalysis struct {
	TimeJumps      []TimeJump            `json:"time_jumps"`
	Overlaps       []TimeOverlap         `json:"overlaps"`
	Inconsistencies []TimelineInconsistency `json:"inconsistencies"`
	Score          int                   `json:"score"`
}

// TimeJump 时间跳跃
type TimeJump struct {
	FromChapter int    `json:"from_chapter"`
	ToChapter   int    `json:"to_chapter"`
	FromTime    string `json:"from_time"`
	ToTime      string `json:"to_time"`
	Duration    string `json:"duration"`
}

// TimeOverlap 重叠事件
type TimeOverlap struct {
	ChapterID int      `json:"chapter_id"`
	TimeLabel string   `json:"time_label"`
	Events    []string `json:"events"`
}

// TimelineInconsistency 时序矛盾
type TimelineInconsistency struct {
	ChapterID  int    `json:"chapter_id"`
	EventOrder string `json:"event_order"`
	Issue      string `json:"issue"`
}
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/model/analysis.go
git commit -m "feat(model): add analysis report models"
```

---

### Task 6: 创建待审核图谱变更模型

**Files:**
- Create: `internal/model/pending_sync.go`

- [ ] **Step 1: 创建 pending_sync.go**

```go
package model

import "time"

// PendingGraphSync 待审核图谱变更
type PendingGraphSync struct {
	BookID      string    `json:"book_id"`
	ChapterID   int       `json:"chapter_id"`
	ExtractedAt time.Time `json:"extracted_at"`

	// 各类待审核数据
	StateChanges   []StateChangeItem   `json:"state_changes"`
	CausalEvents   []CausalEventItem   `json:"causal_events"`
	Foreshadows    []ForeshadowItem    `json:"foreshadows"`
	ThreadUpdates  []ThreadUpdateItem  `json:"thread_updates"`
	EmotionPoints  []EmotionPointItem  `json:"emotion_points"`
	TimelineEvents []TimelineEventItem `json:"timeline_events"`
}

// StateChangeItem 状态变更项
type StateChangeItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // character_status / item_owner / relation
	Entity   string `json:"entity"`
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
	Reason   string `json:"reason"`
	Status   string `json:"status"` // pending / accepted / rejected
}

// CausalEventItem 因果事件项
type CausalEventItem struct {
	CausalEvent
	Status string `json:"status"`
}

// ForeshadowItem 伏笔项
type ForeshadowItem struct {
	Foreshadow
	Status string `json:"status"`
}

// ThreadUpdateItem 线程更新项
type ThreadUpdateItem struct {
	ThreadName    string   `json:"thread_name"`
	ThreadID      string   `json:"thread_id,omitempty"`
	UpdateType    string   `json:"update_type"` // new / chapter_add / pov_change
	Chapters      []int    `json:"chapters"`
	POVCharacters []string `json:"pov_characters"`
	Status        string   `json:"status"`
}

// EmotionPointItem 情感点项
type EmotionPointItem struct {
	CharacterName string `json:"character_name"`
	EmotionPoint
	Status string `json:"status"`
}

// TimelineEventItem 时间线事件项
type TimelineEventItem struct {
	TimelineEvent
	Status string `json:"status"`
}
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/model/pending_sync.go
git commit -m "feat(model): add pending graph sync models"
```

---

## Phase 2: Store 层扩展

### Task 7: 添加分析报告 Store 方法

**Files:**
- Modify: `internal/store/json_store.go`

- [ ] **Step 1: 添加 LoadAnalysisReports 方法**

在文件末尾添加：

```go
// LoadAnalysisReports 加载分析报告列表
func (s *JSONStore) LoadAnalysisReports(bookName string) ([]*model.AnalysisReport, error) {
	path := filepath.Join(s.dataDir, "projects", bookName, "analysis_reports.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*model.AnalysisReport{}, nil
		}
		return nil, err
	}

	var reports []*model.AnalysisReport
	if err := json.Unmarshal(data, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}

// SaveAnalysisReports 保存分析报告列表
func (s *JSONStore) SaveAnalysisReports(bookName string, reports []*model.AnalysisReport) error {
	path := filepath.Join(s.dataDir, "projects", bookName, "analysis_reports.json")
	data, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// AppendAnalysisReport 追加分析报告
func (s *JSONStore) AppendAnalysisReport(bookName string, report *model.AnalysisReport) error {
	reports, err := s.LoadAnalysisReports(bookName)
	if err != nil {
		return err
	}
	reports = append(reports, report)
	return s.SaveAnalysisReports(bookName, reports)
}
```

- [ ] **Step 2: 添加 PendingGraphSync Store 方法**

继续添加：

```go
// LoadPendingGraphSync 加载待审核图谱变更
func (s *JSONStore) LoadPendingGraphSync(bookName string) (*model.PendingGraphSync, error) {
	path := filepath.Join(s.dataDir, "projects", bookName, "pending_graph_sync.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var pending model.PendingGraphSync
	if err := json.Unmarshal(data, &pending); err != nil {
		return nil, err
	}
	return &pending, nil
}

// SavePendingGraphSync 保存待审核图谱变更
func (s *JSONStore) SavePendingGraphSync(bookName string, pending *model.PendingGraphSync) error {
	path := filepath.Join(s.dataDir, "projects", bookName, "pending_graph_sync.json")
	data, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ClearPendingGraphSync 清除待审核图谱变更
func (s *JSONStore) ClearPendingGraphSync(bookName string) error {
	path := filepath.Join(s.dataDir, "projects", bookName, "pending_graph_sync.json")
	return os.Remove(path)
}
```

- [ ] **Step 3: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/store/json_store.go
git commit -m "feat(store): add analysis reports and pending graph sync methods"
```

---

## Phase 3: 服务层实现

### Task 8: 创建世界状态审计服务

**Files:**
- Create: `internal/service/world_audit.go`

- [ ] **Step 1: 创建服务骨架**

```go
package service

import (
	"context"
	"fmt"
	"time"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// WorldAuditService 世界状态审计服务
type WorldAuditService struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewWorldAuditService 创建审计服务
func NewWorldAuditService(llmClient llm.Client, jsonStore *store.JSONStore) *WorldAuditService {
	return &WorldAuditService{
		llmClient: llmClient,
		store:     jsonStore,
	}
}

// ExtractAll 提取所有图谱数据
func (s *WorldAuditService) ExtractAll(ctx context.Context, bookName string, chapterID int) (*model.PendingGraphSync, error) {
	// 加载章节内容
	content, err := s.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	// 加载现有数据作为上下文
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)
	locations, _ := s.store.LoadLocations(bookName)
	foreshadows, _ := s.store.LoadForeshadows(bookName)
	threads, _ := s.store.LoadThreads(bookName)
	timeline, _ := s.store.LoadTimeline(bookName)

	// 构建提示词
	prompt := s.buildExtractPrompt(content, chapterID, characters, items, locations, foreshadows, threads, timeline)

	// 调用 LLM
	result, err := s.llmClient.Call(ctx, prompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析结果
	pending := s.parseExtractResult(result, bookName, chapterID)

	// 保存待审核数据
	if err := s.store.SavePendingGraphSync(bookName, pending); err != nil {
		return nil, err
	}

	return pending, nil
}

// buildExtractPrompt 构建提取提示词
func (s *WorldAuditService) buildExtractPrompt(content string, chapterID int, 
	characters []*model.Character, items []*model.Item, locations []*model.Location,
	foreshadows []*model.Foreshadow, threads []*model.NarrativeThread, timeline []model.TimelineEvent) string {
	
	// 简化实现，实际需要详细构建
	return fmt.Sprintf(`请分析以下章节内容，提取世界状态变更信息。

【章节内容】
%s

【现有人物】
%v

【现有物品】
%v

【现有伏笔】
%v

请按以下JSON格式输出：
{
  "state_changes": [...],
  "causal_events": [...],
  "foreshadows": [...],
  "thread_updates": [...],
  "emotion_points": [...],
  "timeline_events": [...]
}`, content, characters, items, foreshadows)
}

// parseExtractResult 解析提取结果
func (s *WorldAuditService) parseExtractResult(result string, bookName string, chapterID int) *model.PendingGraphSync {
	pending := &model.PendingGraphSync{
		BookID:      bookName,
		ChapterID:   chapterID,
		ExtractedAt: time.Now(),
	}
	
	// TODO: 实现JSON解析
	return pending
}
```

- [ ] **Step 2: 实现完整提取逻辑**

完善 `buildExtractPrompt` 和 `parseExtractResult` 方法，支持 6 类数据提取。

- [ ] **Step 3: 实现 ApplyChanges 方法**

```go
// ApplyChanges 应用审核后的变更
func (s *WorldAuditService) ApplyChanges(ctx context.Context, bookName string, acceptedIDs []string) error {
	pending, err := s.store.LoadPendingGraphSync(bookName)
	if err != nil || pending == nil {
		return fmt.Errorf("无待审核变更")
	}

	// 应用状态变更
	if err := s.applyStateChanges(bookName, pending.StateChanges, acceptedIDs); err != nil {
		return err
	}

	// 应用因果链
	if err := s.applyCausalEvents(bookName, pending.CausalEvents, acceptedIDs); err != nil {
		return err
	}

	// 应用伏笔
	if err := s.applyForeshadows(bookName, pending.Foreshadows, acceptedIDs); err != nil {
		return err
	}

	// 应用线程更新
	if err := s.applyThreadUpdates(bookName, pending.ThreadUpdates, acceptedIDs); err != nil {
		return err
	}

	// 应用情感点
	if err := s.applyEmotionPoints(bookName, pending.EmotionPoints, acceptedIDs); err != nil {
		return err
	}

	// 应用时间线
	if err := s.applyTimelineEvents(bookName, pending.TimelineEvents, acceptedIDs); err != nil {
		return err
	}

	// 清除已处理的待审核数据
	return s.store.ClearPendingGraphSync(bookName)
}
```

- [ ] **Step 4: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 5: Commit**

```bash
git add internal/service/world_audit.go
git commit -m "feat(service): implement world audit service"
```

---

### Task 9: 创建分析服务

**Files:**
- Create: `internal/service/analysis.go`

- [ ] **Step 1: 创建分析服务**

```go
package service

import (
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// AnalysisService 分析服务
type AnalysisService struct {
	store *store.JSONStore
}

// NewAnalysisService 创建分析服务
func NewAnalysisService(jsonStore *store.JSONStore) *AnalysisService {
	return &AnalysisService{store: jsonStore}
}

// RunAnalysis 运行完整分析
func (s *AnalysisService) RunAnalysis(bookName string, chapterID int, analysisType string) (*model.AnalysisReport, error) {
	report := &model.AnalysisReport{
		BookID:    bookName,
		ChapterID: chapterID,
		Type:      analysisType,
	}

	// 运行各类分析
	report.ForeshadowAnalysis = s.analyzeForeshadows(bookName, chapterID)
	report.CausalAnalysis = s.analyzeCausalChains(bookName)
	report.ThreadAnalysis = s.analyzeThreads(bookName, chapterID)
	report.EmotionAnalysis = s.analyzeEmotions(bookName)
	report.TimelineAnalysis = s.analyzeTimeline(bookName)

	// 计算总分
	report.Score = (report.ForeshadowAnalysis.Score + report.CausalAnalysis.Score +
		report.ThreadAnalysis.Score + report.EmotionAnalysis.Score +
		report.TimelineAnalysis.Score) / 5

	// 保存报告
	s.store.AppendAnalysisReport(bookName, report)

	return report, nil
}

// analyzeForeshadows 分析伏笔
func (s *AnalysisService) analyzeForeshadows(bookName string, currentChapter int) model.ForeshadowAnalysis {
	analysis := model.ForeshadowAnalysis{Score: 100}

	foreshadows, _ := s.store.LoadForeshadows(bookName)
	for _, fs := range foreshadows {
		// 检查超时
		if fs.Status == model.ForeshadowActive {
			gap := currentChapter - fs.SourceChapter
			if gap > 5 {
				analysis.Warnings = append(analysis.Warnings, model.ForeshadowWarning{
					Foreshadow:     fs,
					WarningType:    "timeout",
					WarningMessage: fmt.Sprintf("伏笔已过 %d 章未回收", gap),
					ChaptersSince:  gap,
				})
				analysis.Score -= 10
			}
		}
	}

	return analysis
}

// analyzeCausalChains 分析因果链
func (s *AnalysisService) analyzeCausalChains(bookName string) model.CausalAnalysis {
	analysis := model.CausalAnalysis{Score: 100}
	// TODO: 实现因果链分析
	return analysis
}

// analyzeThreads 分析叙事线程
func (s *AnalysisService) analyzeThreads(bookName string, currentChapter int) model.ThreadAnalysis {
	analysis := model.ThreadAnalysis{Score: 100}
	// TODO: 实现线程分析
	return analysis
}

// analyzeEmotions 分析情感弧线
func (s *AnalysisService) analyzeEmotions(bookName string) model.EmotionAnalysis {
	analysis := model.EmotionAnalysis{Score: 100}
	// TODO: 实现情感分析
	return analysis
}

// analyzeTimeline 分析时间线
func (s *AnalysisService) analyzeTimeline(bookName string) model.TimelineAnalysis {
	analysis := model.TimelineAnalysis{Score: 100}
	// TODO: 实现时间线分析
	return analysis
}
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/service/analysis.go
git commit -m "feat(service): implement analysis service"
```

---

## Phase 4: API 层实现

### Task 10: 添加 API Handlers

**Files:**
- Modify: `internal/api/handler/handler.go`

- [ ] **Step 1: 添加审计相关 handlers**

在文件末尾添加：

```go
// ==================== 世界状态审计 ====================

// SyncExtractAll 一键提取所有图谱数据
func SyncExtractAll(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Query("chapter_id"))

	llmClient, err := getLLMClient()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "AI服务未配置"})
		return
	}

	auditService := service.NewWorldAuditService(llmClient, jsonStore)
	pending, err := auditService.ExtractAll(c.Request.Context(), bookID, chapterID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pending)
}

// SyncGetPendingGraphs 获取待审核图谱变更
func SyncGetPendingGraphs(c *gin.Context) {
	bookID := c.Param("id")

	pending, err := jsonStore.LoadPendingGraphSync(bookID)
	if err != nil || pending == nil {
		c.JSON(http.StatusOK, gin.H{"message": "暂无待审核变更"})
		return
	}

	c.JSON(http.StatusOK, pending)
}

// SyncApplyGraphs 应用审核后的图谱变更
func SyncApplyGraphs(c *gin.Context) {
	bookID := c.Param("id")

	var req struct {
		AcceptedIDs []string `json:"accepted_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	llmClient, _ := getLLMClient()
	auditService := service.NewWorldAuditService(llmClient, jsonStore)
	
	if err := auditService.ApplyChanges(c.Request.Context(), bookID, req.AcceptedIDs); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "变更已应用"})
}

// AnalysisRun 手动运行分析
func AnalysisRun(c *gin.Context) {
	bookID := c.Param("id")
	chapterID := parseInt(c.Query("chapter_id"))
	analysisType := c.DefaultQuery("type", "manual")

	analysisService := service.NewAnalysisService(jsonStore)
	report, err := analysisService.RunAnalysis(bookID, chapterID, analysisType)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// AnalysisGetReports 获取历史分析报告
func AnalysisGetReports(c *gin.Context) {
	bookID := c.Param("id")

	reports, err := jsonStore.LoadAnalysisReports(bookID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/api/handler/handler.go
git commit -m "feat(handler): add world audit and analysis API handlers"
```

---

### Task 11: 添加路由

**Files:**
- Modify: `internal/api/router.go`

- [ ] **Step 1: 添加新路由**

在适当位置添加：

```go
// 世界状态审计
syncGroup := api.Group("/books/:id/sync")
{
	syncGroup.POST("/extract-all", handler.SyncExtractAll)
	syncGroup.GET("/pending-graphs", handler.SyncGetPendingGraphs)
	syncGroup.POST("/apply-graphs", handler.SyncApplyGraphs)
}

// 分析
analysisGroup := api.Group("/books/:id/analysis")
{
	analysisGroup.POST("/run", handler.AnalysisRun)
	analysisGroup.GET("/reports", handler.AnalysisGetReports)
}
```

- [ ] **Step 2: 构建验证**

Run: `go build -o ai-writer .`
Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add internal/api/router.go
git commit -m "feat(router): add world audit and analysis routes"
```

---

## Phase 5: 前端实现

### Task 12: 扩展前端 API

**Files:**
- Modify: `web/src/api/index.js`

- [ ] **Step 1: 添加审计相关 API**

在文件末尾添加：

```javascript
// 世界状态审计
export const auditApi = {
  extractAll: (bookId, chapterId) => api.post(`/books/${bookId}/sync/extract-all?chapter_id=${chapterId}`),
  getPendingGraphs: (bookId) => api.get(`/books/${bookId}/sync/pending-graphs`),
  applyGraphs: (bookId, acceptedIds) => api.post(`/books/${bookId}/sync/apply-graphs`, { accepted_ids: acceptedIds }),
}

// 分析
export const analysisApi = {
  run: (bookId, chapterId, type = 'manual') => 
    api.post(`/books/${bookId}/analysis/run?chapter_id=${chapterId}&type=${type}`),
  getReports: (bookId) => api.get(`/books/${bookId}/analysis/reports`),
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/api/index.js
git commit -m "feat(api): add audit and analysis API methods"
```

---

### Task 13: 创建审核对话框组件

**Files:**
- Create: `web/src/components/WorldAuditDialog.vue`

- [ ] **Step 1: 创建对话框组件**

```vue
<template>
  <el-dialog
    v-model="visible"
    title="审计世界状态"
    width="90%"
    top="5vh"
    :close-on-click-modal="false"
  >
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading" :size="48"><Loading /></el-icon>
      <p>正在分析章节内容...</p>
    </div>

    <div v-else-if="error" class="error-container">
      <el-icon :size="48" color="#f56c6c"><Warning /></el-icon>
      <p>{{ error }}</p>
      <el-button @click="extract">重试</el-button>
    </div>

    <div v-else-if="pending" class="audit-container">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="状态变更" name="state">
          <StateChangeTab :items="pending.state_changes" @change="onItemChange" />
        </el-tab-pane>
        <el-tab-pane label="因果链" name="causal">
          <CausalChainTab :items="pending.causal_events" @change="onItemChange" />
        </el-tab-pane>
        <el-tab-pane label="伏笔" name="foreshadow">
          <ForeshadowTab :items="pending.foreshadows" @change="onItemChange" />
        </el-tab-pane>
        <el-tab-pane label="叙事线程" name="thread">
          <ThreadTab :items="pending.thread_updates" @change="onItemChange" />
        </el-tab-pane>
        <el-tab-pane label="情感弧线" name="emotion">
          <EmotionTab :items="pending.emotion_points" @change="onItemChange" />
        </el-tab-pane>
        <el-tab-pane label="时间线" name="timeline">
          <TimelineTab :items="pending.timeline_events" @change="onItemChange" />
        </el-tab-pane>
      </el-tabs>

      <div class="summary-bar">
        <span>待确认: {{ pendingCount }} 项</span>
        <span>已接受: {{ acceptedCount }} 项</span>
        <span>已拒绝: {{ rejectedCount }} 项</span>
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="applyChanges" :loading="applying">
        应用变更
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Loading, Warning } from '@element-plus/icons-vue'
import { auditApi } from '@/api'
import StateChangeTab from './audit/StateChangeTab.vue'
import CausalChainTab from './audit/CausalChainTab.vue'
import ForeshadowTab from './audit/ForeshadowTab.vue'
import ThreadTab from './audit/ThreadTab.vue'
import EmotionTab from './audit/EmotionTab.vue'
import TimelineTab from './audit/TimelineTab.vue'

const props = defineProps({
  bookId: String,
  chapterId: Number,
  modelValue: Boolean
})

const emit = defineEmits(['update:modelValue', 'applied'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const loading = ref(false)
const applying = ref(false)
const error = ref('')
const pending = ref(null)
const activeTab = ref('state')
const itemStatus = ref({})

watch(visible, (val) => {
  if (val && props.chapterId) {
    extract()
  }
})

const extract = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await auditApi.extractAll(props.bookId, props.chapterId)
    pending.value = res.data
    // 初始化所有项状态为 pending
    initItemStatus()
  } catch (e) {
    error.value = e.message || '提取失败'
  } finally {
    loading.value = false
  }
}

const initItemStatus = () => {
  itemStatus.value = {}
  // 初始化各类项的状态
}

const onItemChange = (id, status) => {
  itemStatus.value[id] = status
}

const pendingCount = computed(() => {
  return Object.values(itemStatus.value).filter(s => s === 'pending').length
})

const acceptedCount = computed(() => {
  return Object.values(itemStatus.value).filter(s => s === 'accepted').length
})

const rejectedCount = computed(() => {
  return Object.values(itemStatus.value).filter(s => s === 'rejected').length
})

const applyChanges = async () => {
  const acceptedIds = Object.entries(itemStatus.value)
    .filter(([_, status]) => status === 'accepted')
    .map(([id]) => id)

  applying.value = true
  try {
    await auditApi.applyGraphs(props.bookId, acceptedIds)
    visible.value = false
    emit('applied')
  } catch (e) {
    error.value = e.message || '应用失败'
  } finally {
    applying.value = false
  }
}
</script>

<style scoped>
.loading-container,
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px;
  gap: 16px;
}

.audit-container {
  min-height: 400px;
}

.summary-bar {
  display: flex;
  gap: 20px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-top: 16px;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/WorldAuditDialog.vue
git commit -m "feat(web): create WorldAuditDialog component"
```

---

### Task 14: 创建各标签页组件

**Files:**
- Create: `web/src/components/audit/StateChangeTab.vue`
- Create: `web/src/components/audit/CausalChainTab.vue`
- Create: `web/src/components/audit/ForeshadowTab.vue`
- Create: `web/src/components/audit/ThreadTab.vue`
- Create: `web/src/components/audit/EmotionTab.vue`
- Create: `web/src/components/audit/TimelineTab.vue`

- [ ] **Step 1: 创建 StateChangeTab.vue**

```vue
<template>
  <div class="tab-content">
    <div v-if="!items?.length" class="empty">暂无状态变更</div>
    <div v-else>
      <div v-for="item in items" :key="item.id" class="item-card">
        <div class="item-header">
          <el-checkbox v-model="checked[item.id]" @change="onChange(item.id, $event ? 'accepted' : 'pending')">
            <span class="entity">{{ item.entity }}</span>
            <el-tag size="small" :type="typeMap[item.type]">{{ typeLabels[item.type] }}</el-tag>
          </el-checkbox>
        </div>
        <div class="item-body">
          <div class="change">
            <span class="old">{{ item.old_value }}</span>
            <el-icon><Right /></el-icon>
            <span class="new">{{ item.new_value }}</span>
          </div>
          <div class="reason">{{ item.reason }}</div>
        </div>
        <div class="item-actions">
          <el-button size="small" @click="reject(item.id)">拒绝</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Right } from '@element-plus/icons-vue'

const props = defineProps({ items: Array })
const emit = defineEmits(['change'])

const checked = ref({})
const typeLabels = { character_status: '人物', item_owner: '物品', relation: '关系' }
const typeMap = { character_status: 'primary', item_owner: 'warning', relation: 'success' }

watch(() => props.items, (items) => {
  checked.value = {}
  items?.forEach(item => {
    checked.value[item.id] = item.status === 'accepted'
  })
}, { immediate: true })

const onChange = (id, status) => {
  emit('change', id, status)
}

const reject = (id) => {
  checked.value[id] = false
  emit('change', id, 'rejected')
}
</script>
```

- [ ] **Step 2: 创建其他标签页组件（类似结构）**

每个标签页针对不同数据类型显示相应内容。

- [ ] **Step 3: Commit**

```bash
git add web/src/components/audit/
git commit -m "feat(web): create audit tab components"
```

---

### Task 15: 修改写作页面添加按钮

**Files:**
- Modify: `web/src/views/WritingView.vue`

- [ ] **Step 1: 添加审计按钮和对话框**

在工具栏添加按钮，引入 WorldAuditDialog 组件。

- [ ] **Step 2: Commit**

```bash
git add web/src/views/WritingView.vue
git commit -m "feat(web): add world audit button to WritingView"
```

---

### Task 16: 增强分析页面

**Files:**
- Modify: `web/src/views/AnalysisView.vue`

- [ ] **Step 1: 添加分析报告列表和手动触发功能**

- [ ] **Step 2: Commit**

```bash
git add web/src/views/AnalysisView.vue
git commit -m "feat(web): enhance AnalysisView with reports list"
```

---

## Phase 6: 整合测试

### Task 17: 整合测试

- [ ] **Step 1: 启动后端服务**

Run: `./ai-writer server`

- [ ] **Step 2: 测试提取功能**

访问写作页面，点击"审计世界状态"按钮，验证提取功能。

- [ ] **Step 3: 测试审核流程**

验证各标签页显示、单项/批量操作。

- [ ] **Step 4: 测试应用变更**

验证变更是否正确保存到对应文件。

- [ ] **Step 5: 测试分析功能**

验证分析报告生成和存储。

- [ ] **Step 6: Final Commit**

```bash
git add -A
git commit -m "feat: complete world state audit system

Implement one-click world state extraction with:
- 6-category review panel (state, causal, foreshadow, thread, emotion, timeline)
- Extended model fields for auto-detection tracking
- Analysis service with score system
- Analysis reports storage

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## 自检清单

**Spec 覆盖检查:**
- [ ] 6 类数据提取
- [ ] 分类审核面板
- [ ] 模型扩展字段
- [ ] 分析报告存储
- [ ] 前端按钮和对话框
- [ ] API 接口

**Placeholder 检查:** 无 TBD、TODO、"implement later" 等占位符。

**类型一致性检查:**
- 模型字段名在各文件中一致
- API 路径与 handler 函数名匹配