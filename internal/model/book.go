package model

import "time"

// Book 书籍/项目
type Book struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联数据（运行时加载）
	Volumes    []*Volume    `json:"volumes,omitempty"`
	Chapters   []*Chapter   `json:"chapters,omitempty"`
	Characters []*Character `json:"characters,omitempty"`
	Items      []*Item      `json:"items,omitempty"`
	Locations  []*Location  `json:"locations,omitempty"`
}

// BookMeta 书籍元数据（存储在 metadata.json）
type BookMeta struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Genre       string    `json:"genre,omitempty"`
	TargetWords int       `json:"target_words,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 架构师数据
	ArchitectData *ArchitectData `json:"architect_data,omitempty"`
}

// ArchitectData 架构师生成数据
type ArchitectData struct {
	Synopsis       *SynopsisJSON                `json:"synopsis,omitempty"`
	WorldView      *WorldViewJSON               `json:"world_view,omitempty"`
	Volumes        []VolumeJSON                 `json:"volumes,omitempty"`
	ChapterDetails map[string]ChapterDetailJSON `json:"chapter_details,omitempty"` // 章节细纲，key 为 "卷索引_章索引"
	CurrentStep    int                          `json:"current_step,omitempty"`    // 当前步骤进度
	UpdatedAt      time.Time                    `json:"updated_at"`
}

// SynopsisJSON 总纲数据（前端格式）
type SynopsisJSON struct {
	Title       string   `json:"title"`
	Genre       string   `json:"genre"`
	Theme       string   `json:"theme"`
	WordCount   int      `json:"word_count"`
	Synopsis    string   `json:"synopsis"`
	MainPlot    string   `json:"main_plot"`
	SubPlots    []string `json:"sub_plots"`
	MainChars   []string `json:"main_chars"`
	EndingType  string   `json:"ending_type"`
	VolumeCount int      `json:"volume_count"`
}

// WorldViewJSON 世界观数据（前端格式）
type WorldViewJSON struct {
	Genre           string `json:"genre"`
	Era             string `json:"era"`
	TechLevel       string `json:"tech_level"`
	PowerSystem     string `json:"power_system"`
	SocialStructure string `json:"social_structure"`
	SpecialRules    string `json:"special_rules"`
	ImportantItems  string `json:"important_items"`
	Organizations   string `json:"organizations"`
	Locations       string `json:"locations"`
	History         string `json:"history"`
	MainConflict    string `json:"main_conflict"`
	Development     string `json:"development"`
}

// VolumeJSON 分卷数据（前端格式）
type VolumeJSON struct {
	ID           string          `json:"id"`
	Index        int             `json:"index"`
	Title        string          `json:"title"`
	Synopsis     string          `json:"synopsis"`
	MainEvent    string          `json:"main_event"`
	EmotionArc   string          `json:"emotion_arc"`
	ChapterCount int             `json:"chapter_count"`
	Chapters     []ChapterJSON   `json:"chapters"`
}

// ChapterJSON 章节大纲数据
type ChapterJSON struct {
	ID         string `json:"id"`
	Index      int    `json:"index"`
	VolumeIndex int   `json:"volume_index"`
	Title      string `json:"title"`
	Synopsis   string `json:"synopsis"`
	MainEvent  string `json:"main_event"`
	Characters string `json:"characters"`
	Location   string `json:"location"`
	Foreshadow string `json:"foreshadow"`
}

// ChapterDetailJSON 章节细纲数据
type ChapterDetailJSON struct {
	ChapterKey  string        `json:"chapter_key"` // 如 "0_1" 表示第一卷第二章
	WordTarget  int           `json:"word_target"`
	Scenes      []SceneJSON   `json:"scenes,omitempty"`
	Dialogues   []string      `json:"dialogues,omitempty"`
	Foreshadows []string      `json:"foreshadows,omitempty"`
}

// SceneJSON 场景数据
type SceneJSON struct {
	Location   string `json:"location"`
	Characters string `json:"characters"`
	Event      string `json:"event"`
	Mood       string `json:"mood"`
}

// Volume 分卷
type Volume struct {
	ID          string `json:"id"`
	BookID      string `json:"book_id"`
	Title       string `json:"title"`
	Order       int    `json:"order"`
	Description string `json:"description,omitempty"`
}

// Chapter 章节
type Chapter struct {
	ID        int    `json:"id"`
	BookID    string `json:"book_id"`
	VolumeID  string `json:"volume_id"`
	Title     string `json:"title"`
	Outline   string `json:"outline"`
	Summary   string `json:"summary,omitempty"`
	WordCount int    `json:"word_count"`

	// 时间信息
	TimeInfo TimeInfo `json:"time_info"`

	// 审稿结果
	ReviewData *ReviewData `json:"review_data,omitempty"`

	// 因果链（新增）
	CausalChain *CausalEvent `json:"causal_chain,omitempty"`

	// 叙事线程（新增）
	ThreadID string `json:"thread_id,omitempty"`

	// 出场实体索引（自动提取）
	Characters []string `json:"characters,omitempty"` // 出场人物名称列表
	Items      []string `json:"items,omitempty"`      // 出场物品名称列表
	Locations  []string `json:"locations,omitempty"`  // 出场地点名称列表

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TimeInfo 时间信息
type TimeInfo struct {
	Label    string   `json:"label"`    // 如 "修仙历10年春"
	Duration string   `json:"duration"` // 如 "3天"
	Events   []string `json:"events"`   // 本章发生的事件
}

// ReviewData 审稿结果
type ReviewData struct {
	OverallScore int           `json:"overall_score"`
	Issues       []ReviewIssue `json:"issues"`
	ReviewedAt   time.Time     `json:"reviewed_at"`
	Fixed        bool          `json:"fixed"`
}

// ReviewIssue 审稿问题
type ReviewIssue struct {
	Type        string `json:"type"`        // 人设/逻辑/节奏/文笔
	Severity    string `json:"severity"`    // 严重/中等/轻微
	Location    string `json:"location"`    // 问题位置
	Description string `json:"description"` // 问题描述
	Suggestion  string `json:"suggestion"`  // 修改建议
}

// Paragraph 段落
type Paragraph struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	WordCount int    `json:"word_count"`
}

// ChapterParagraphs 章节段落结构
type ChapterParagraphs struct {
	ChapterID  int         `json:"chapter_id"`
	Paragraphs []Paragraph `json:"paragraphs"`
	Metadata   ParagraphMetadata `json:"metadata"`
}

// ParagraphMetadata 段落元数据
type ParagraphMetadata struct {
	ParagraphCount int       `json:"paragraph_count"`
	TotalWords     int       `json:"total_words"`
	UpdatedAt      time.Time `json:"updated_at"`
}