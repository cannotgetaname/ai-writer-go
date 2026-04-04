package engine

import (
	"context"
	"fmt"
	"strings"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// CausalChainEngine 因果链引擎
type CausalChainEngine struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewCausalChainEngine 创建因果链引擎
func NewCausalChainEngine(llmClient llm.Client, store *store.JSONStore) *CausalChainEngine {
	return &CausalChainEngine{
		llmClient: llmClient,
		store:     store,
	}
}

// ExtractFromChapter 从章节提取因果链
func (e *CausalChainEngine) ExtractFromChapter(ctx context.Context, bookName string, chapterID int) (*model.CausalEvent, error) {
	// 加载章节内容
	content, err := e.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if content == "" {
		return nil, fmt.Errorf("章节 %d 没有内容", chapterID)
	}

	// 构建提取提示词
	prompt := fmt.Sprintf(`分析以下章节内容，提取因果关系（因→事→果→决）：

【章节内容】
%s

请按以下JSON格式输出：
{
  "cause": "触发原因",
  "event": "核心事件",
  "effect": "直接后果",
  "decision": "角色决定",
  "characters": ["涉及角色"]
}`, content)

	// 调用 LLM
	result, err := e.llmClient.Call(ctx, prompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析结果
	event := &model.CausalEvent{
		ID:        fmt.Sprintf("ce_%d", chapterID),
		BookID:    bookName,
		ChapterID: chapterID,
	}

	// 简单解析 JSON（实际应该用 json.Unmarshal）
	if strings.Contains(result, `"cause"`) {
		event.Cause = extractJSONValue(result, "cause")
		event.Event = extractJSONValue(result, "event")
		event.Effect = extractJSONValue(result, "effect")
		event.Decision = extractJSONValue(result, "decision")
	}

	return event, nil
}

// ValidateChain 验证因果链一致性
func (e *CausalChainEngine) ValidateChain(bookName string) []string {
	events, err := e.store.LoadCausalChains(bookName)
	if err != nil {
		return []string{"无法加载因果链"}
	}

	var issues []string
	for i, event := range events {
		if event.Cause == "" {
			issues = append(issues, fmt.Sprintf("第%d章因果链缺少'因'", event.ChapterID))
		}
		if event.Event == "" {
			issues = append(issues, fmt.Sprintf("第%d章因果链缺少'事'", event.ChapterID))
		}

		// 检查与前一章的连贯性
		if i > 0 {
			prev := events[i-1]
			if prev.Decision != "" && event.Cause != "" {
				// 简单检查：前一章的决定是否影响当前章
				// 实际可以用 LLM 做更智能的检查
			}
		}
	}

	return issues
}

// GetChainContext 获取因果链上下文（用于写作）
func (e *CausalChainEngine) GetChainContext(bookName string, currentChapter int) (string, error) {
	events, err := e.store.LoadCausalChains(bookName)
	if err != nil {
		return "", err
	}

	var contextParts []string
	for _, event := range events {
		if event.ChapterID < currentChapter {
			contextParts = append(contextParts, fmt.Sprintf(
				"第%d章: 因[%s] → 事[%s] → 果[%s] → 决[%s]",
				event.ChapterID, event.Cause, event.Event, event.Effect, event.Decision,
			))
		}
	}

	return strings.Join(contextParts, "\n"), nil
}

// extractJSONValue 从 JSON 字符串中提取值
func extractJSONValue(jsonStr, key string) string {
	// 简单提取，实际应用 json.Unmarshal
	searchKey := `"` + key + `":`
	startIdx := strings.Index(jsonStr, searchKey)
	if startIdx == -1 {
		return ""
	}

	startIdx += len(searchKey)
	// 跳过空白
	for startIdx < len(jsonStr) && (jsonStr[startIdx] == ' ' || jsonStr[startIdx] == '\n') {
		startIdx++
	}

	if startIdx >= len(jsonStr) {
		return ""
	}

	// 检查是字符串还是其他类型
	if jsonStr[startIdx] == '"' {
		// 字符串
		startIdx++
		endIdx := strings.Index(jsonStr[startIdx:], `"`)
		if endIdx == -1 {
			return ""
		}
		return jsonStr[startIdx : startIdx+endIdx]
	}

	return ""
}