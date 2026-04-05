package model

// Item 物品
type Item struct {
	ID          string `json:"id"`
	BookID      string `json:"book_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`        // 武器/法宝/丹药/材料
	Owner       string `json:"owner"`       // 持有者
	Description string `json:"description"` // 描述
	Origin      string `json:"origin,omitempty"`  // 来历
	Abilities   string `json:"abilities,omitempty"` // 能力

	// 新增字段
	Rank    string `json:"rank,omitempty"`    // 品阶/等级 (如: 天阶、地阶、玄阶、黄阶)
	Faction string `json:"faction,omitempty"` // 所属势力
	Sect    string `json:"sect,omitempty"`    // 所属宗门
	Location string `json:"location,omitempty"` // 所在地点（物品可能存放在某地点而非人物持有）

	// 章节追踪
	AppearChapters []int `json:"appear_chapters,omitempty"` // 出场章节列表

	// 归属历史
	OwnerHistory []ItemOwnerChange `json:"owner_history,omitempty"`
}

// ItemOwnerChange 物品归属变更记录
type ItemOwnerChange struct {
	ChapterID int    `json:"chapter_id"`
	OldOwner  string `json:"old_owner"`
	NewOwner  string `json:"new_owner"`
	Action    string `json:"action"` // 获得/失去/赠送/抢夺
	Reason    string `json:"reason"`
}

// Location 地点
type Location struct {
	ID          string   `json:"id"`
	BookID      string   `json:"book_id"`
	Name        string   `json:"name"`
	Parent      string   `json:"parent,omitempty"`  // 父级地点
	Neighbors   []string `json:"neighbors,omitempty"` // 相邻地点
	Description string   `json:"description"`
	Faction     string   `json:"faction,omitempty"` // 所属势力
	Danger      string   `json:"danger,omitempty"`  // 危险等级
}