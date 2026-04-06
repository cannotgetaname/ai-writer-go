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