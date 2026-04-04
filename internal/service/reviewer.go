package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-writer/internal/config"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// ReviewService 审稿服务
type ReviewService struct {
	llmClient llm.Client
	store     *store.JSONStore
	prompts   *config.PromptsConfig
}

// NewReviewService 创建审稿服务
func NewReviewService(llmClient llm.Client, store *store.JSONStore, prompts *config.PromptsConfig) *ReviewService {
	return &ReviewService{
		llmClient: llmClient,
		store:     store,
		prompts:   prompts,
	}
}

// ReviewResult 审稿结果
type ReviewResult struct {
	OverallScore int            `json:"overall_score"`
	Issues       []ReviewIssue  `json:"issues"`
	Suggestions  []string       `json:"suggestions"`
}

// ReviewIssue 审稿问题
type ReviewIssue struct {
	Type        string `json:"type"`        // 人设/逻辑/节奏/文笔
	Severity    string `json:"severity"`    // 严重/中等/轻微
	Location    string `json:"location"`    // 问题位置
	Description string `json:"description"` // 问题描述
	Suggestion  string `json:"suggestion"`  // 修改建议
}

// ReviewChapter 审稿章节
func (s *ReviewService) ReviewChapter(ctx context.Context, bookName string, chapterID int) (*ReviewResult, error) {
	// 加载章节内容
	content, err := s.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if content == "" {
		return nil, fmt.Errorf("章节 %d 没有内容", chapterID)
	}

	// 加载人物设定
	characters, _ := s.store.LoadCharacters(bookName)

	// 构建审稿提示词
	userPrompt := s.buildReviewPrompt(content, characters)

	// 调用 LLM
	systemPrompt := s.prompts.ReviewerSystem
	result, err := s.llmClient.CallWithSystem(ctx, systemPrompt, userPrompt, "reviewer")
	if err != nil {
		return nil, err
	}

	// 解析结果
	reviewResult := s.parseReviewResult(result)

	return reviewResult, nil
}

// AuditChapter 状态审计章节
func (s *ReviewService) AuditChapter(ctx context.Context, bookName string, chapterID int) (*model.Chapter, error) {
	// 加载章节内容
	content, err := s.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if content == "" {
		return nil, fmt.Errorf("章节 %d 没有内容", chapterID)
	}

	// 加载现有状态
	chapters, _ := s.store.LoadChapters(bookName)
	characters, _ := s.store.LoadCharacters(bookName)
	items, _ := s.store.LoadItems(bookName)
	locations, _ := s.store.LoadLocations(bookName)

	// 构建审计提示词
	userPrompt := s.buildAuditPrompt(content, characters, items, locations)

	// 调用 LLM
	systemPrompt := s.prompts.AuditorSystem
	result, err := s.llmClient.CallWithSystem(ctx, systemPrompt, userPrompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析审计结果并应用变更
	auditResult := s.parseAuditResult(result)
	if auditResult != nil {
		s.applyAuditChanges(bookName, chapterID, auditResult)
	}

	// 返回更新后的章节
	for _, ch := range chapters {
		if ch.ID == chapterID {
			return ch, nil
		}
	}

	return nil, fmt.Errorf("章节 %d 不存在", chapterID)
}

// buildReviewPrompt 构建审稿提示词
func (s *ReviewService) buildReviewPrompt(content string, characters []*model.Character) string {
	var prompt strings.Builder

	prompt.WriteString("请对以下章节内容进行审稿：\n\n")
	prompt.WriteString("【章节内容】\n")
	prompt.WriteString(content)
	prompt.WriteString("\n\n")

	if len(characters) > 0 {
		prompt.WriteString("【人物设定】\n")
		for _, char := range characters {
			prompt.WriteString(fmt.Sprintf("- %s(%s): %s\n", char.Name, char.Role, char.Bio))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString(`请从以下维度进行审查：
1. 【人设一致性】人物言行是否符合其性格和身份？
2. 【剧情逻辑】是否有前后矛盾或不合理的转折？
3. 【爽点节奏】是否过于拖沓？是否有期待感？
4. 【文笔表达】描写是否生动？对话是否自然？

请输出 Markdown 格式的审稿报告。`)

	return prompt.String()
}

// buildAuditPrompt 构建审计提示词
func (s *ReviewService) buildAuditPrompt(content string, characters []*model.Character, items []*model.Item, locations []*model.Location) string {
	var prompt strings.Builder

	prompt.WriteString("请分析以下章节内容，提取状态变更：\n\n")
	prompt.WriteString("【章节内容】\n")
	prompt.WriteString(content)
	prompt.WriteString("\n\n")

	if len(characters) > 0 {
		prompt.WriteString("【现有人物】\n")
		for _, char := range characters {
			prompt.WriteString(fmt.Sprintf("- %s: 状态=%s\n", char.Name, char.Status))
		}
		prompt.WriteString("\n")
	}

	if len(items) > 0 {
		prompt.WriteString("【现有物品】\n")
		for _, item := range items {
			prompt.WriteString(fmt.Sprintf("- %s: 持有者=%s\n", item.Name, item.Owner))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString(`请提取以下信息（JSON 格式）：
{
  "character_changes": [
    {"name": "角色名", "status": "新状态", "reason": "原因"}
  ],
  "item_changes": [
    {"name": "物品名", "owner": "新持有者", "action": "获得/失去"}
  ],
  "new_characters": ["新出现的角色"],
  "new_items": ["新出现的物品"],
  "new_locations": ["新出现的地点"],
  "time_progression": "时间进展描述"
}`)

	return prompt.String()
}

// parseReviewResult 解析审稿结果
func (s *ReviewService) parseReviewResult(result string) *ReviewResult {
	review := &ReviewResult{
		OverallScore: 70, // 默认分数
		Issues:       []ReviewIssue{},
		Suggestions:  []string{},
	}

	// 尝试提取分数
	score := s.extractScore(result)
	if score > 0 {
		review.OverallScore = score
	}

	// 解析问题列表
	issues := s.extractIssues(result)
	review.Issues = issues

	// 提取建议
	suggestions := s.extractSuggestions(result)
	review.Suggestions = suggestions

	return review
}

// extractScore 从审稿报告中提取分数
func (s *ReviewService) extractScore(result string) int {
	// 尝试匹配各种分数格式
	patterns := []string{
		"综合评分", "总体评分", "总评", "得分", "分数", "评分",
	}

	for _, pattern := range patterns {
		idx := strings.Index(result, pattern)
		if idx == -1 {
			continue
		}

		// 在关键词后查找数字
		subStr := result[idx:]
		for i := len(pattern); i < len(subStr) && i < 30; i++ {
			c := subStr[i]
			if c >= '0' && c <= '9' {
				// 提取完整数字
				numStart := i
				for i < len(subStr) && subStr[i] >= '0' && subStr[i] <= '9' {
					i++
				}
				numStr := subStr[numStart:i]
				var score int
				fmt.Sscanf(numStr, "%d", &score)
				if score >= 0 && score <= 100 {
					return score
				}
			}
		}
	}

	// 尝试匹配 "X/100" 或 "X/10" 格式
	for i := 0; i < len(result); i++ {
		if result[i] == '/' {
			// 检查后面是否是 100 或 10
			if i+3 < len(result) && result[i+1:i+4] == "100" {
				// 向前找数字
				numEnd := i
				numStart := numEnd
				for numStart > 0 && result[numStart-1] >= '0' && result[numStart-1] <= '9' {
					numStart--
				}
				if numStart < numEnd {
					var score int
					fmt.Sscanf(result[numStart:numEnd], "%d", &score)
					if score >= 0 && score <= 100 {
						return score
					}
				}
			}
		}
	}

	return 0
}

// extractIssues 从审稿报告中提取问题列表
func (s *ReviewService) extractIssues(result string) []ReviewIssue {
	var issues []ReviewIssue

	lines := strings.Split(result, "\n")
	var currentIssue *ReviewIssue

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检测问题标题行（如 "1.", "- ", "【人设】" 等）
		if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") ||
			strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") {
			if currentIssue != nil {
				issues = append(issues, *currentIssue)
			}
			currentIssue = &ReviewIssue{
				Description: strings.TrimPrefix(line, string(line[0])+". "),
			}

			// 尝试识别类型
			if strings.Contains(line, "人设") || strings.Contains(line, "人物") {
				currentIssue.Type = "人设"
			} else if strings.Contains(line, "逻辑") || strings.Contains(line, "剧情") {
				currentIssue.Type = "逻辑"
			} else if strings.Contains(line, "节奏") || strings.Contains(line, "爽点") {
				currentIssue.Type = "节奏"
			} else if strings.Contains(line, "文笔") || strings.Contains(line, "描写") {
				currentIssue.Type = "文笔"
			}

			// 尝试识别严重程度
			if strings.Contains(line, "严重") || strings.Contains(line, "重要") {
				currentIssue.Severity = "严重"
			} else if strings.Contains(line, "中等") || strings.Contains(line, "一般") {
				currentIssue.Severity = "中等"
			} else {
				currentIssue.Severity = "轻微"
			}

		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// 可能是建议或问题详情
			if currentIssue != nil {
				if currentIssue.Suggestion == "" {
					currentIssue.Suggestion = strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
				}
			} else {
				// 独立问题
				issues = append(issues, ReviewIssue{
					Type:        "其他",
					Description: strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "),
					Severity:    "中等",
				})
			}
		} else if strings.HasPrefix(line, "建议") || strings.HasPrefix(line, "修改建议") {
			if currentIssue != nil {
				currentIssue.Suggestion = strings.TrimPrefix(line, "建议：")
				currentIssue.Suggestion = strings.TrimPrefix(currentIssue.Suggestion, "建议:")
			}
		}
	}

	// 添加最后一个问题
	if currentIssue != nil {
		issues = append(issues, *currentIssue)
	}

	return issues
}

// extractSuggestions 从审稿报告中提取建议
func (s *ReviewService) extractSuggestions(result string) []string {
	var suggestions []string

	lines := strings.Split(result, "\n")
	inSuggestionSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检测建议部分开始
		if strings.Contains(line, "修改建议") || strings.Contains(line, "改进建议") ||
			strings.Contains(line, "建议") {
			inSuggestionSection = true
			continue
		}

		// 在建议部分内提取建议
		if inSuggestionSection {
			if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") ||
				strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") {
				suggestion := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
				suggestion = strings.TrimPrefix(strings.TrimPrefix(suggestion, "1. "), "2. ")
				if len(suggestion) > 5 {
					suggestions = append(suggestions, suggestion)
				}
			} else if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "##") {
				// 新的部分开始，结束建议部分
				if len(suggestions) > 0 {
					break
				}
			}
		}
	}

	return suggestions
}

// AuditResult 审计结果
type AuditResult struct {
	CharacterChanges []CharacterChange `json:"character_changes"`
	ItemChanges       []ItemChange      `json:"item_changes"`
	NewCharacters     []NewCharacter   `json:"new_characters"`
	NewItems          []NewItem        `json:"new_items"`
	NewLocations      []NewLocation    `json:"new_locations"`
	TimeProgression   string           `json:"time_progression"`
}

// CharacterChange 人物状态变更
type CharacterChange struct {
	Name   string `json:"name"`
	Field  string `json:"field"`
	Old    string `json:"old"`
	New    string `json:"new"`
	Reason string `json:"reason"`
}

// ItemChange 物品变更
type ItemChange struct {
	Name      string `json:"name"`
	OldOwner  string `json:"old_owner"`
	NewOwner  string `json:"new_owner"`
	Reason    string `json:"reason"`
}

// NewCharacter 新人物
type NewCharacter struct {
	Name string `json:"name"`
	Role string `json:"role"`
	Bio  string `json:"bio"`
}

// NewItem 新物品
type NewItem struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Owner string `json:"owner"`
}

// NewLocation 新地点
type NewLocation struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// parseAuditResult 解析审计结果
func (s *ReviewService) parseAuditResult(result string) *AuditResult {
	// 尝试解析 JSON
	var data AuditResult

	if err := parseJSONResult(result, &data); err != nil {
		// JSON 解析失败，尝试文本解析
		return s.parseAuditText(result)
	}

	return &data
}

// parseAuditText 从文本解析审计结果
func (s *ReviewService) parseAuditText(result string) *AuditResult {
	audit := &AuditResult{
		CharacterChanges: []CharacterChange{},
		ItemChanges:      []ItemChange{},
		NewCharacters:    []NewCharacter{},
		NewItems:         []NewItem{},
		NewLocations:     []NewLocation{},
	}

	lines := strings.Split(result, "\n")
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检测部分标题
		if strings.Contains(line, "人物状态") || strings.Contains(line, "角色状态") {
			currentSection = "character"
		} else if strings.Contains(line, "物品") {
			currentSection = "item"
		} else if strings.Contains(line, "新人物") || strings.Contains(line, "新角色") {
			currentSection = "new_character"
		} else if strings.Contains(line, "新物品") {
			currentSection = "new_item"
		} else if strings.Contains(line, "新地点") {
			currentSection = "new_location"
		} else if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			// 解析列表项
			item := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")

			switch currentSection {
			case "character":
				// 格式: 角色名: 状态变更
				if parts := strings.SplitN(item, ":", 2); len(parts) == 2 {
					audit.CharacterChanges = append(audit.CharacterChanges, CharacterChange{
						Name: strings.TrimSpace(parts[0]),
						New:  strings.TrimSpace(parts[1]),
					})
				}
			case "item":
				if parts := strings.SplitN(item, ":", 2); len(parts) == 2 {
					audit.ItemChanges = append(audit.ItemChanges, ItemChange{
						Name: strings.TrimSpace(parts[0]),
						NewOwner: strings.TrimSpace(parts[1]),
					})
				}
			case "new_character":
				if parts := strings.SplitN(item, ":", 2); len(parts) == 2 {
					audit.NewCharacters = append(audit.NewCharacters, NewCharacter{
						Name: strings.TrimSpace(parts[0]),
						Bio:  strings.TrimSpace(parts[1]),
					})
				}
			case "new_item":
				if parts := strings.SplitN(item, ":", 2); len(parts) == 2 {
					audit.NewItems = append(audit.NewItems, NewItem{
						Name: strings.TrimSpace(parts[0]),
						Type: strings.TrimSpace(parts[1]),
					})
				}
			case "new_location":
				audit.NewLocations = append(audit.NewLocations, NewLocation{
					Name: item,
				})
			}
		}
	}

	return audit
}

// applyAuditChanges 应用审计变更
func (s *ReviewService) applyAuditChanges(bookName string, chapterID int, audit *AuditResult) {
	// 应用人物状态变更
	if len(audit.CharacterChanges) > 0 {
		characters, _ := s.store.LoadCharacters(bookName)
		for _, change := range audit.CharacterChanges {
			for _, char := range characters {
				if char.Name == change.Name {
					switch change.Field {
					case "status", "":
						char.Status = change.New
					}
				}
			}
		}
		s.store.SaveCharacters(bookName, characters)
	}

	// 应用物品变更
	if len(audit.ItemChanges) > 0 {
		items, _ := s.store.LoadItems(bookName)
		for _, change := range audit.ItemChanges {
			for _, item := range items {
				if item.Name == change.Name {
					item.Owner = change.NewOwner
				}
			}
		}
		s.store.SaveItems(bookName, items)
	}

	// 添加新人物
	if len(audit.NewCharacters) > 0 {
		characters, _ := s.store.LoadCharacters(bookName)
		for _, nc := range audit.NewCharacters {
			// 检查是否已存在
			exists := false
			for _, c := range characters {
				if c.Name == nc.Name {
					exists = true
					break
				}
			}
			if !exists {
				characters = append(characters, &model.Character{
					ID:     generateID(),
					BookID: bookName,
					Name:   nc.Name,
					Role:   nc.Role,
					Bio:    nc.Bio,
					Status: "存活",
				})
			}
		}
		s.store.SaveCharacters(bookName, characters)
	}

	// 添加新物品
	if len(audit.NewItems) > 0 {
		items, _ := s.store.LoadItems(bookName)
		for _, ni := range audit.NewItems {
			exists := false
			for _, i := range items {
				if i.Name == ni.Name {
					exists = true
					break
				}
			}
			if !exists {
				items = append(items, &model.Item{
					ID:     generateID(),
					BookID: bookName,
					Name:   ni.Name,
					Type:   ni.Type,
					Owner:  ni.Owner,
				})
			}
		}
		s.store.SaveItems(bookName, items)
	}

	// 添加新地点
	if len(audit.NewLocations) > 0 {
		locations, _ := s.store.LoadLocations(bookName)
		for _, nl := range audit.NewLocations {
			exists := false
			for _, l := range locations {
				if l.Name == nl.Name {
					exists = true
					break
				}
			}
			if !exists {
				locations = append(locations, &model.Location{
					ID:     generateID(),
					BookID: bookName,
					Name:   nl.Name,
				})
			}
		}
		s.store.SaveLocations(bookName, locations)
	}

	// 更新时间信息
	if audit.TimeProgression != "" {
		chapters, _ := s.store.LoadChapters(bookName)
		for _, ch := range chapters {
			if ch.ID == chapterID {
				ch.TimeInfo.Label = audit.TimeProgression
				break
			}
		}
		s.store.SaveChapters(bookName, chapters)
	}
}

// parseJSONResult 解析 JSON 辅助函数
func parseJSONResult(input string, v interface{}) error {
	// 尝试找到 JSON 部分
	start := strings.Index(input, "{")
	if start == -1 {
		return fmt.Errorf("未找到 JSON")
	}

	end := strings.LastIndex(input, "}")
	if end == -1 || end < start {
		return fmt.Errorf("JSON 格式不完整")
	}

	jsonStr := input[start : end+1]
	return json.Unmarshal([]byte(jsonStr), v)
}