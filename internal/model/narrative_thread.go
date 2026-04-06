package model

import "time"

// ThreadType 叙事线程类型
type ThreadType string

const (
	ThreadMain      ThreadType = "main"      // 主线
	ThreadSub       ThreadType = "sub"       // 支线
	ThreadParallel  ThreadType = "parallel"  // 并行线
	ThreadFlashback ThreadType = "flashback" // 闪回线
)

// ThreadStatus 线程状态
type ThreadStatus string

const (
	ThreadActive   ThreadStatus = "active"
	ThreadPaused   ThreadStatus = "paused"
	ThreadComplete ThreadStatus = "complete"
)

// NarrativeThread 叙事线程
type NarrativeThread struct {
	ID     string      `json:"id"`
	BookID string      `json:"book_id"`
	Name   string      `json:"name"`
	Type   ThreadType  `json:"type"`

	// 视角角色
	POVCharacters []string `json:"pov_characters"`

	// 目标弧线
	Goal         string `json:"goal"`
	StartChapter int    `json:"start_chapter"`
	EndChapter   int    `json:"end_chapter,omitempty"`

	// 篇幅权重
	Weight int `json:"weight"` // 字数分配权重

	// 状态
	Status            ThreadStatus `json:"status"`
	LastActiveChapter int          `json:"last_active_chapter"`

	// 关联章节
	Chapters []int `json:"chapters"`

	// 新增字段
	AutoDetected bool `json:"auto_detected"`

	CreatedAt time.Time `json:"created_at"`
}

// TimelineEvent 时间线事件
type TimelineEvent struct {
	ChapterID   int      `json:"chapter_id"`
	ThreadID    string   `json:"thread_id,omitempty"`
	TimeLabel   string   `json:"time_label"`
	Duration    string   `json:"duration"`
	Events      []string `json:"events"`
	Characters  []string `json:"characters,omitempty"`
	Location    string   `json:"location,omitempty"`

	// 新增字段
	AutoDetected bool `json:"auto_detected"`
}
