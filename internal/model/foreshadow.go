package model

import "time"

// ForeshadowType 伏笔类型
type ForeshadowType string

const (
	ForeshadowTypeItem     ForeshadowType = "item"      // 物品伏笔
	ForeshadowTypeCharacter ForeshadowType = "character" // 人物伏笔
	ForeshadowTypePlot     ForeshadowType = "plot"      // 剧情伏笔
	ForeshadowTypeMystery  ForeshadowType = "mystery"   // 悬念
	ForeshadowTypeSetting  ForeshadowType = "setting"   // 设定伏笔
	ForeshadowTypePromise  ForeshadowType = "promise"   // 承诺
	ForeshadowTypeConflict ForeshadowType = "conflict"  // 冲突
)

// ForeshadowStatus 伏笔状态
type ForeshadowStatus string

const (
	ForeshadowActive    ForeshadowStatus = "active"    // 埋设中
	ForeshadowResolved  ForeshadowStatus = "resolved"  // 已回收
	ForeshadowExpired   ForeshadowStatus = "expired"   // 过期预警
	ForeshadowAbandoned ForeshadowStatus = "abandoned" // 已放弃
)

// Importance 重要程度
type Importance string

const (
	ImportanceHigh   Importance = "high"
	ImportanceMedium Importance = "medium"
	ImportanceLow    Importance = "low"
)

// Foreshadow 伏笔
type Foreshadow struct {
	ID     string          `json:"id"`
	BookID string          `json:"book_id"`
	Type   ForeshadowType  `json:"type"`
	Content string         `json:"content"`      // 伏笔内容
	Importance Importance  `json:"importance"`   // 重要程度

	// 埋设信息
	SourceChapter   int    `json:"source_chapter"`
	SourceParagraph string `json:"source_paragraph,omitempty"`

	// 预期回收
	TargetChapter int `json:"target_chapter,omitempty"`

	// 回收信息
	Status          ForeshadowStatus `json:"status"`
	ResolvedChapter int              `json:"resolved_chapter,omitempty"`
	ResolvedContent string           `json:"resolved_content,omitempty"`

	// 关联因果链
	CausalEventID string `json:"causal_event_id,omitempty"`

	Notes string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ForeshadowWarning 伏笔预警
type ForeshadowWarning struct {
	Foreshadow    *Foreshadow `json:"foreshadow"`
	WarningType   string      `json:"warning_type"`   // overdue / target_missed
	WarningMessage string     `json:"warning_message"`
	ChaptersSince int         `json:"chapters_since"`
}

// ForeshadowSettings 伏笔设置
type ForeshadowSettings struct {
	WarningChapterThreshold int  `json:"warning_chapter_threshold"` // 预警阈值
	AutoDetectEnabled       bool `json:"auto_detect_enabled"`       // 审稿时自动检测
}

// ForeshadowStats 伏笔统计
type ForeshadowStats struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Resolved  int `json:"resolved"`
	Expired   int `json:"expired"`
	Abandoned int `json:"abandoned"`
}