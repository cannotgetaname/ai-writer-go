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