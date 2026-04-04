package model

// WorldView 世界观
type WorldView struct {
	BookID string `json:"book_id"`

	// 基本信息
	BasicInfo WorldViewBasic `json:"basic_info"`

	// 核心设定
	CoreSettings WorldViewCore `json:"core_settings"`

	// 关键元素
	KeyElements WorldViewElements `json:"key_elements"`

	// 背景故事
	Background WorldViewBackground `json:"background"`

	// Markdown 格式（兼容）
	Markdown string `json:"markdown,omitempty"`
}

type WorldViewBasic struct {
	Genre     string `json:"genre"`      // 题材类型
	Era       string `json:"era"`        // 时代背景
	TechLevel string `json:"tech_level"` // 科技水平
}

type WorldViewCore struct {
	PowerSystem     string `json:"power_system"`     // 力量体系
	SocialStructure string `json:"social_structure"` // 社会结构
	SpecialRules    string `json:"special_rules"`    // 特殊规则
}

type WorldViewElements struct {
	ImportantItems string `json:"important_items"` // 重要物品
	Organizations  string `json:"organizations"`   // 势力组织
	Locations      string `json:"locations"`       // 主要地点
}

type WorldViewBackground struct {
	History      string `json:"history"`       // 历史背景
	MainConflict string `json:"main_conflict"` // 主要矛盾
	Development  string `json:"development"`   // 发展趋势
}

// WorldViewTemplate 世界观模板
type WorldViewTemplate struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	BasicInfo    WorldViewBasic     `json:"basic_info"`
	CoreSettings WorldViewCore      `json:"core_settings"`
	KeyElements  WorldViewElements  `json:"key_elements"`
	Background   WorldViewBackground `json:"background"`
}