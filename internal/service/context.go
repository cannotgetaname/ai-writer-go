package service

import (
	"fmt"
	"strings"

	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// ContextService 写作上下文服务
type ContextService struct {
	store *store.JSONStore
}

// NewContextService 创建上下文服务
func NewContextService(jsonStore *store.JSONStore) *ContextService {
	return &ContextService{
		store: jsonStore,
	}
}

// WritingContext 写作上下文
type WritingContext struct {
	// 世界观
	WorldViewSummary string `json:"worldview_summary"`

	// 人物
	Characters []CharacterContext `json:"characters"`

	// 前情
	RecentSummary string `json:"recent_summary"`

	// 伏笔
	ActiveForeshadows []ForeshadowHint `json:"active_foreshadows"`

	// 因果链
	CausalChain string `json:"causal_chain"`

	// 时间
	CurrentTime string `json:"current_time"`

	// 原始数据（用于调试）
	Raw struct {
		WorldView   *model.WorldView     `json:"worldview"`
		Characters  []*model.Character   `json:"characters"`
		Chapters    []*model.Chapter     `json:"chapters"`
		Foreshadows []*model.Foreshadow  `json:"foreshadows"`
		Events      []*model.CausalEvent `json:"events"`
	} `json:"-"`
}

// CharacterContext 人物上下文
type CharacterContext struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Location string `json:"location,omitempty"`
	Bio      string `json:"bio"`
	Emotion  string `json:"emotion,omitempty"`
}

// ForeshadowHint 伏笔提示
type ForeshadowHint struct {
	ID            string `json:"id"`
	Content       string `json:"content"`
	SourceChapter int    `json:"source_chapter"`
	Gap           int    `json:"gap"` // 距当前章节的距离
	Type          string `json:"type"`
}

// BuildContext 构建写作上下文
func (s *ContextService) BuildContext(bookName string, chapterID int, maxContextWords int) (*WritingContext, error) {
	ctx := &WritingContext{}

	// 1. 加载世界观
	worldview, err := s.store.LoadWorldView(bookName)
	if err == nil && worldview != nil {
		ctx.Raw.WorldView = worldview
		ctx.WorldViewSummary = s.summarizeWorldView(worldview, maxContextWords/5)
	}

	// 2. 加载人物
	characters, err := s.store.LoadCharacters(bookName)
	if err == nil {
		ctx.Raw.Characters = characters
		ctx.Characters = s.buildCharacterContexts(characters, chapterID)
	}

	// 3. 加载章节
	chapters, err := s.store.LoadChapters(bookName)
	if err == nil {
		ctx.Raw.Chapters = chapters
		ctx.RecentSummary = s.buildRecentSummary(bookName, chapters, chapterID, maxContextWords/3)
		ctx.CurrentTime = s.getCurrentTime(chapters, chapterID)
	}

	// 4. 加载伏笔
	foreshadows, err := s.store.LoadForeshadows(bookName)
	if err == nil {
		ctx.Raw.Foreshadows = foreshadows
		ctx.ActiveForeshadows = s.buildForeshadowHints(foreshadows, chapterID)
	}

	// 5. 加载因果链
	events, err := s.store.LoadCausalChains(bookName)
	if err == nil {
		ctx.Raw.Events = events
		ctx.CausalChain = s.buildCausalChainContext(events, chapterID)
	}

	return ctx, nil
}

// summarizeWorldView 生成世界观摘要
func (s *ContextService) summarizeWorldView(worldview *model.WorldView, maxWords int) string {
	var parts []string

	if worldview.BasicInfo.Genre != "" {
		parts = append(parts, fmt.Sprintf("题材: %s", worldview.BasicInfo.Genre))
	}
	if worldview.CoreSettings.PowerSystem != "" {
		parts = append(parts, fmt.Sprintf("力量体系: %s", worldview.CoreSettings.PowerSystem))
	}
	if worldview.CoreSettings.SocialStructure != "" {
		parts = append(parts, fmt.Sprintf("社会结构: %s", worldview.CoreSettings.SocialStructure))
	}
	if worldview.CoreSettings.SpecialRules != "" {
		parts = append(parts, fmt.Sprintf("特殊规则: %s", worldview.CoreSettings.SpecialRules))
	}
	if worldview.KeyElements.Organizations != "" {
		parts = append(parts, fmt.Sprintf("势力: %s", worldview.KeyElements.Organizations))
	}

	summary := strings.Join(parts, "\n")

	// 截断到最大字数
	if len(summary) > maxWords {
		summary = summary[:maxWords] + "..."
	}

	return summary
}

// buildCharacterContexts 构建人物上下文
func (s *ContextService) buildCharacterContexts(characters []*model.Character, currentChapter int) []CharacterContext {
	var contexts []CharacterContext

	for _, char := range characters {
		// 只包含主要角色
		if char.Role != "主角" && char.Role != "配角" && char.Role != "反派" {
			continue
		}

		cc := CharacterContext{
			Name:   char.Name,
			Role:   char.Role,
			Status: char.Status,
			Bio:    char.Bio,
		}

		// 获取最新情感状态
		if len(char.EmotionalArc) > 0 {
			for i := len(char.EmotionalArc) - 1; i >= 0; i-- {
				if char.EmotionalArc[i].ChapterID < currentChapter {
					cc.Emotion = char.EmotionalArc[i].Emotion
					break
				}
			}
		}

		contexts = append(contexts, cc)
	}

	return contexts
}

// buildRecentSummary 构建前情提要
func (s *ContextService) buildRecentSummary(bookName string, chapters []*model.Chapter, currentChapter int, maxWords int) string {
	// 找到当前章节索引
	currentIndex := -1
	for i, ch := range chapters {
		if ch.ID == currentChapter {
			currentIndex = i
			break
		}
	}

	if currentIndex <= 0 {
		return ""
	}

	var summaries []string

	// 获取前3章摘要
	start := currentIndex - 3
	if start < 0 {
		start = 0
	}

	for i := start; i < currentIndex; i++ {
		ch := chapters[i]

		var summary string
		if ch.Summary != "" {
			summary = ch.Summary
		} else {
			// 尝试加载内容生成简要摘要
			content, err := s.store.LoadChapterContent(bookName, ch.ID)
			if err == nil && len(content) > 100 {
				// 简单截取前100字作为摘要
				summary = content
				if len(summary) > 150 {
					summary = summary[:150] + "..."
				}
			}
		}

		if summary != "" {
			summaries = append(summaries, fmt.Sprintf("第%d章 %s: %s", ch.ID, ch.Title, summary))
		}
	}

	result := strings.Join(summaries, "\n\n")

	// 截断到最大字数
	if len(result) > maxWords {
		result = result[:maxWords] + "..."
	}

	return result
}

// buildForeshadowHints 构建伏笔提示
func (s *ContextService) buildForeshadowHints(foreshadows []*model.Foreshadow, currentChapter int) []ForeshadowHint {
	var hints []ForeshadowHint

	for _, fs := range foreshadows {
		// 只包含活跃伏笔
		if fs.Status != model.ForeshadowActive {
			continue
		}

		gap := currentChapter - fs.SourceChapter

		hint := ForeshadowHint{
			ID:            truncate(fs.ID, 8),
			Content:       fs.Content,
			SourceChapter: fs.SourceChapter,
			Gap:           gap,
			Type:          string(fs.Type),
		}

		hints = append(hints, hint)
	}

	// 只返回最近的10个伏笔
	if len(hints) > 10 {
		hints = hints[:10]
	}

	return hints
}

// buildCausalChainContext 构建因果链上下文
func (s *ContextService) buildCausalChainContext(events []*model.CausalEvent, currentChapter int) string {
	var parts []string

	// 只取前5章的因果链
	startChapter := currentChapter - 5
	if startChapter < 1 {
		startChapter = 1
	}

	for _, event := range events {
		if event.ChapterID >= startChapter && event.ChapterID < currentChapter {
			part := fmt.Sprintf("第%d章: 因[%s] → 事[%s] → 果[%s] → 决[%s]",
				event.ChapterID,
				truncate(event.Cause, 20),
				truncate(event.Event, 20),
				truncate(event.Effect, 20),
				truncate(event.Decision, 20),
			)
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, "\n")
}

// getCurrentTime 获取当前时间
func (s *ContextService) getCurrentTime(chapters []*model.Chapter, currentChapter int) string {
	// 找到当前章节
	for _, ch := range chapters {
		if ch.ID == currentChapter {
			return ch.TimeInfo.Label
		}
	}

	// 找最近的有时间信息的章节
	for i := len(chapters) - 1; i >= 0; i-- {
		if chapters[i].TimeInfo.Label != "" && chapters[i].ID < currentChapter {
			return chapters[i].TimeInfo.Label
		}
	}

	return ""
}

// FormatContext 格式化上下文为提示词
func (s *ContextService) FormatContext(ctx *WritingContext) string {
	var prompt strings.Builder

	// 世界观
	if ctx.WorldViewSummary != "" {
		prompt.WriteString("【世界观】\n")
		prompt.WriteString(ctx.WorldViewSummary)
		prompt.WriteString("\n\n")
	}

	// 人物
	if len(ctx.Characters) > 0 {
		prompt.WriteString("【主要人物】\n")
		for _, char := range ctx.Characters {
			line := fmt.Sprintf("- %s(%s): %s", char.Name, char.Role, char.Bio)
			if char.Status != "存活" {
				line += fmt.Sprintf(" [状态: %s]", char.Status)
			}
			if char.Emotion != "" {
				line += fmt.Sprintf(" [情绪: %s]", char.Emotion)
			}
			prompt.WriteString(line + "\n")
		}
		prompt.WriteString("\n")
	}

	// 前情
	if ctx.RecentSummary != "" {
		prompt.WriteString("【前情提要】\n")
		prompt.WriteString(ctx.RecentSummary)
		prompt.WriteString("\n\n")
	}

	// 活跃伏笔
	if len(ctx.ActiveForeshadows) > 0 {
		prompt.WriteString("【活跃伏笔】\n")
		for _, fs := range ctx.ActiveForeshadows {
			prompt.WriteString(fmt.Sprintf("- [%s] %s (第%d章埋设)\n", fs.ID, fs.Content, fs.SourceChapter))
		}
		prompt.WriteString("\n")
	}

	// 因果链
	if ctx.CausalChain != "" {
		prompt.WriteString("【因果链】\n")
		prompt.WriteString(ctx.CausalChain)
		prompt.WriteString("\n\n")
	}

	// 时间
	if ctx.CurrentTime != "" {
		prompt.WriteString(fmt.Sprintf("【当前时间】%s\n\n", ctx.CurrentTime))
	}

	return prompt.String()
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}