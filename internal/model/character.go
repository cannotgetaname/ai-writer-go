package model

import "time"

// DramaticaRole Dramatica 角色职能
type DramaticaRole string

const (
	RoleProtagonist  DramaticaRole = "protagonist"  // 主角
	RoleAntagonist   DramaticaRole = "antagonist"   // 反派
	RoleImpact       DramaticaRole = "impact"       // 冲击者
	RoleGuardian     DramaticaRole = "guardian"     // 守护者
	RoleContagonist  DramaticaRole = "contagonist"  // 阻碍者
	RoleSidekick     DramaticaRole = "sidekick"     // 伙伴
	RoleSkeptic      DramaticaRole = "skeptic"      // 怀疑者
)

// InfoSource 信息来源
type InfoSource string

const (
	SourceWitnessed InfoSource = "witnessed" // 亲眼所见
	SourceHearsay   InfoSource = "hearsay"   // 道听途说
	SourceDeduced   InfoSource = "deduced"   // 推理得出
	SourceDocument  InfoSource = "document"  // 文献记载
)

// Character 人物
type Character struct {
	ID        string `json:"id"`
	BookID    string `json:"book_id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"` // 男/女/未知
	Role      string `json:"role"`   // 主角/配角/反派/路人
	Status    string `json:"status"` // 存活/死亡/失踪
	Bio       string `json:"bio"`    // 简介
	Avatar    string `json:"avatar,omitempty"`

	// 势力与身份
	Faction     string `json:"faction,omitempty"`     // 所属势力/组织
	Sect        string `json:"sect,omitempty"`        // 宗门/门派
	Position    string `json:"position,omitempty"`    // 职位/身份
	Cultivation string `json:"cultivation,omitempty"` // 修为境界

	// Dramatica 扩展字段
	DramaticaRole  DramaticaRole `json:"dramatica_role"`           // Dramatica 职能
	ExternalGoal   string        `json:"external_goal,omitempty"`  // 外部目标（可见）
	InternalDesire string        `json:"internal_desire,omitempty"` // 内在渴望（不自知）

	// 关系
	Relations []Relation `json:"relations"`

	// 情感弧线（新增）
	EmotionalArc []EmotionPoint `json:"emotional_arc,omitempty"`

	// 信息边界（新增）
	KnownInfos []KnownInfo `json:"known_infos,omitempty"`

	// 状态历史
	StatusHistory []StatusChange `json:"status_history,omitempty"`

	// 章节追踪
	AppearChapters []int `json:"appear_chapters,omitempty"` // 出场章节列表

	// 势力变更历史
	FactionHistory []FactionChange `json:"faction_history,omitempty"`
}

// FactionChange 势力变更记录
type FactionChange struct {
	ChapterID  int    `json:"chapter_id"`
	OldFaction string `json:"old_faction"`
	NewFaction string `json:"new_faction"`
	Reason     string `json:"reason"`
}

// Relation 人物关系
type Relation struct {
	TargetID   string            `json:"target_id"`
	TargetName string            `json:"target_name"`
	Type       string            `json:"type"`   // 朋友/敌人/师徒/恋人
	Value      int               `json:"value"`  // -100 到 +100
	History    []RelationChange  `json:"history,omitempty"`
}

// RelationChange 关系变化记录
type RelationChange struct {
	ChapterID int    `json:"chapter_id"`
	Change    int    `json:"change"` // 变化值 +20, -30 等
	Reason    string `json:"reason"` // 变化原因
}

// EmotionPoint 情感点
type EmotionPoint struct {
	ChapterID int    `json:"chapter_id"`
	Emotion   string `json:"emotion"`   // 情绪类型：愤怒/悲伤/喜悦
	Intensity int    `json:"intensity"` // 1-10 强度
	Trigger   string `json:"trigger"`   // 触发事件
}

// KnownInfo 已知信息（信息边界）
type KnownInfo struct {
	ID             string     `json:"id"`
	InfoKey        string     `json:"info_key"`       // 信息标识
	Content        string     `json:"content"`        // 信息内容
	LearnedChapter int        `json:"learned_chapter"`
	Source         InfoSource `json:"source"` // 信息来源
}

// StatusChange 状态变化
type StatusChange struct {
	ChapterID int       `json:"chapter_id"`
	Field     string    `json:"field"`     // 状态字段
	OldValue  string    `json:"old_value"` // 旧值
	NewValue  string    `json:"new_value"` // 新值
	Reason    string    `json:"reason"`    // 变化原因
	ChangedAt time.Time `json:"changed_at"`
}