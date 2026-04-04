package model

import "time"

// CausalStatus 因果链状态
type CausalStatus string

const (
	CausalActive    CausalStatus = "active"    // 进行中
	CausalResolved  CausalStatus = "resolved"  // 已解决
	CausalAbandoned CausalStatus = "abandoned" // 已放弃
)

// CausalEvent 因果事件
type CausalEvent struct {
	ID        string       `json:"id"`
	BookID    string       `json:"book_id"`
	ChapterID int          `json:"chapter_id"`

	// 核心因果结构
	Cause      string   `json:"cause"`      // 因：触发原因
	Event      string   `json:"event"`      // 事：核心事件
	Effect     string   `json:"effect"`     // 果：直接后果
	Decision   string   `json:"decision"`   // 决：角色决定

	// 涉及角色
	Characters []string `json:"characters"`

	// 关联伏笔
	ForeshadowIDs []string `json:"foreshadow_ids,omitempty"`

	// 状态
	Status CausalStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

// CausalLink 因果链接
type CausalLink struct {
	FromEventID string `json:"from_event_id"` // 前因事件
	ToEventID   string `json:"to_event_id"`   // 后果事件
	LinkType    string `json:"link_type"`     // leads_to / enables / blocks
}

// CausalChain 因果链（章节级别）
type CausalChain struct {
	BookID string         `json:"book_id"`
	Events []*CausalEvent `json:"events"`
	Links  []*CausalLink  `json:"links,omitempty"`
}