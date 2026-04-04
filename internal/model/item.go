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