package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// ConsistencyEngine 一致性检查引擎
type ConsistencyEngine struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewConsistencyEngine 创建一致性检查引擎
func NewConsistencyEngine(llmClient llm.Client, store *store.JSONStore) *ConsistencyEngine {
	return &ConsistencyEngine{
		llmClient: llmClient,
		store:     store,
	}
}

// ConsistencyIssue 一致性问题
type ConsistencyIssue struct {
	Type        string `json:"type"`        // character/item/location/plot/time
	ChapterFrom int    `json:"chapter_from"` // 问题起始章节
	ChapterTo   int    `json:"chapter_to"`   // 问题结束章节
	Entity      string `json:"entity"`       // 相关实体
	Description string `json:"description"`  // 问题描述
	Severity    string `json:"severity"`     // high/medium/low
	Suggestion  string `json:"suggestion"`  // 修改建议
}

// ConsistencyReport 一致性检查报告
type ConsistencyReport struct {
	BookName     string            `json:"book_name"`
	ChapterRange string            `json:"chapter_range"`
	Issues       []ConsistencyIssue `json:"issues"`
	Summary      string            `json:"summary"`
}

// CheckChapter 检查章节一致性
func (e *ConsistencyEngine) CheckChapter(ctx context.Context, bookName string, chapterID int) (*ConsistencyReport, error) {
	// 加载章节内容
	content, err := e.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if content == "" {
		return nil, fmt.Errorf("章节 %d 没有内容", chapterID)
	}

	// 加载项目设定作为参照
	characters, _ := e.store.LoadCharacters(bookName)
	items, _ := e.store.LoadItems(bookName)
	locations, _ := e.store.LoadLocations(bookName)
	worldview, _ := e.store.LoadWorldView(bookName)
	chapters, _ := e.store.LoadChapters(bookName)

	// 构建检查提示词
	prompt := e.buildCheckPrompt(content, chapterID, characters, items, locations, worldview, chapters)

	// 调用 LLM
	result, err := e.llmClient.Call(ctx, prompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析结果
	report := &ConsistencyReport{
		BookName:     bookName,
		ChapterRange: fmt.Sprintf("%d", chapterID),
		Issues:       e.parseIssues(result),
		Summary:      e.extractSummary(result),
	}

	return report, nil
}

// CheckRange 检查章节范围的一致性
func (e *ConsistencyEngine) CheckRange(ctx context.Context, bookName string, fromChapter, toChapter int) (*ConsistencyReport, error) {
	var allIssues []ConsistencyIssue

	for chapterID := fromChapter; chapterID <= toChapter; chapterID++ {
		report, err := e.CheckChapter(ctx, bookName, chapterID)
		if err != nil {
			continue // 跳过无法检查的章节
		}

		allIssues = append(allIssues, report.Issues...)
	}

	// 生成汇总报告
	report := &ConsistencyReport{
		BookName:     bookName,
		ChapterRange: fmt.Sprintf("%d-%d", fromChapter, toChapter),
		Issues:       allIssues,
		Summary:      e.generateSummary(allIssues),
	}

	return report, nil
}

// buildCheckPrompt 构建检查提示词
func (e *ConsistencyEngine) buildCheckPrompt(content string, chapterID int, characters []*model.Character, items []*model.Item, locations []*model.Location, worldview *model.WorldView, chapters []*model.Chapter) string {
	var prompt strings.Builder

	prompt.WriteString("请检查以下章节内容与项目设定的一致性：\n\n")

	// 章节内容
	prompt.WriteString("【章节内容】\n")
	if len(content) > 3000 {
		prompt.WriteString(content[:3000] + "...")
	} else {
		prompt.WriteString(content)
	}
	prompt.WriteString("\n\n")

	// 人物设定
	if len(characters) > 0 {
		prompt.WriteString("【人物设定】\n")
		for _, char := range characters {
			prompt.WriteString(fmt.Sprintf("- %s(%s): 状态=%s, 简介=%s\n", char.Name, char.Role, char.Status, char.Bio))
		}
		prompt.WriteString("\n")
	}

	// 物品设定
	if len(items) > 0 {
		prompt.WriteString("【物品设定】\n")
		for _, item := range items {
			prompt.WriteString(fmt.Sprintf("- %s: 类型=%s, 持有者=%s\n", item.Name, item.Type, item.Owner))
		}
		prompt.WriteString("\n")
	}

	// 地点设定
	if len(locations) > 0 {
		prompt.WriteString("【地点设定】\n")
		for _, loc := range locations {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", loc.Name, loc.Description))
		}
		prompt.WriteString("\n")
	}

	// 世界观
	if worldview != nil {
		prompt.WriteString("【世界观】\n")
		prompt.WriteString(fmt.Sprintf("类型: %s\n", worldview.BasicInfo.Genre))
		if worldview.CoreSettings.PowerSystem != "" {
			prompt.WriteString(fmt.Sprintf("力量体系: %s\n", worldview.CoreSettings.PowerSystem))
		}
		prompt.WriteString("\n")
	}

	// 时间线
	if len(chapters) > 0 {
		prompt.WriteString("【时间线】\n")
		for _, ch := range chapters {
			if ch.ID <= chapterID && ch.TimeInfo.Label != "" {
				prompt.WriteString(fmt.Sprintf("第%d章: %s\n", ch.ID, ch.TimeInfo.Label))
			}
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString(`请检查以下一致性问题，用JSON格式输出：
{
  "issues": [
    {
      "type": "character/item/location/plot/time",
      "entity": "相关实体名称",
      "description": "问题描述",
      "severity": "high/medium/low",
      "suggestion": "修改建议"
    }
  ],
  "summary": "整体评价"
}

检查要点：
1. 人物言行是否符合人设
2. 物品归属是否正确
3. 地点描述是否一致
4. 剧情是否前后矛盾
5. 时间线是否合理

如果没有发现问题，返回空数组。`)

	return prompt.String()
}

// parseIssues 解析问题列表
func (e *ConsistencyEngine) parseIssues(result string) []ConsistencyIssue {
	var issues []ConsistencyIssue

	// 尝试解析 JSON
	var data struct {
		Issues []struct {
			Type        string `json:"type"`
			Entity      string `json:"entity"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			Suggestion  string `json:"suggestion"`
		} `json:"issues"`
	}

	if err := parseJSON(result, &data); err != nil {
		// JSON 解析失败，尝试简单文本解析
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "问题") || strings.Contains(line, "矛盾") || strings.Contains(line, "不一致") {
				issues = append(issues, ConsistencyIssue{
					Type:        "unknown",
					Description: line,
					Severity:    "medium",
				})
			}
		}
		return issues
	}

	for _, issue := range data.Issues {
		issues = append(issues, ConsistencyIssue{
			Type:        issue.Type,
			Entity:      issue.Entity,
			Description: issue.Description,
			Severity:    issue.Severity,
			Suggestion:  issue.Suggestion,
		})
	}

	return issues
}

// extractSummary 提取摘要
func (e *ConsistencyEngine) extractSummary(result string) string {
	var data struct {
		Summary string `json:"summary"`
	}

	if err := parseJSON(result, &data); err != nil {
		return "一致性检查完成"
	}

	return data.Summary
}

// generateSummary 生成汇总摘要
func (e *ConsistencyEngine) generateSummary(issues []ConsistencyIssue) string {
	if len(issues) == 0 {
		return "未发现一致性问题"
	}

	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, issue := range issues {
		switch issue.Severity {
		case "high":
			highCount++
		case "medium":
			mediumCount++
		default:
			lowCount++
		}
	}

	return fmt.Sprintf("发现 %d 个问题（严重:%d, 中等:%d, 轻微:%d）", len(issues), highCount, mediumCount, lowCount)
}

// parseJSON 解析 JSON（辅助函数）
func parseJSON(input string, v interface{}) error {
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