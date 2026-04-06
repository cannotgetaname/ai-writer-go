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
	ThreadIDs    []string `json:"thread_ids"`
	ChapterID    int      `json:"chapter_id"`
	ConflictType string   `json:"conflict_type"`
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