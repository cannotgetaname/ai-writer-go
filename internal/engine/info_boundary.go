package engine

import (
	"context"
	"fmt"
	"strings"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// InfoBoundaryManager 信息边界管理器
type InfoBoundaryManager struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewInfoBoundaryManager 创建信息边界管理器
func NewInfoBoundaryManager(llmClient llm.Client, store *store.JSONStore) *InfoBoundaryManager {
	return &InfoBoundaryManager{
		llmClient: llmClient,
		store:     store,
	}
}

// CheckInfoLeak 检测信息越界
func (m *InfoBoundaryManager) CheckInfoLeak(ctx context.Context, bookName string, chapterID int) ([]string, error) {
	// 加载章节内容
	content, err := m.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	// 加载人物
	characters, err := m.store.LoadCharacters(bookName)
	if err != nil {
		return nil, err
	}

	var leaks []string

	for _, char := range characters {
		// 获取该角色在当前章节之前已知的信息
		knownInfos := m.getKnownInfosBeforeChapter(char, chapterID)

		if len(knownInfos) == 0 {
			continue
		}

		// 使用 LLM 检测是否有信息越界
		prompt := fmt.Sprintf(`分析以下章节内容，检测角色"%s"是否使用了不该知道的信息。

【角色已知信息】
%s

【章节内容】
%s

请指出角色是否有不符合其已知信息的言行，如有请说明。如无问题请回复"无问题"。`,
			char.Name, strings.Join(knownInfos, "\n"), content)

		result, err := m.llmClient.Call(ctx, prompt, "auditor")
		if err != nil {
			continue
		}

		if !strings.Contains(result, "无问题") {
			leaks = append(leaks, fmt.Sprintf("[%s] %s", char.Name, result))
		}
	}

	return leaks, nil
}

// UpdateKnownInfo 更新角色已知信息
func (m *InfoBoundaryManager) UpdateKnownInfo(char *model.Character, info *model.KnownInfo) {
	char.KnownInfos = append(char.KnownInfos, *info)
}

// getKnownInfosBeforeChapter 获取章节之前的已知信息
func (m *InfoBoundaryManager) getKnownInfosBeforeChapter(char *model.Character, chapterID int) []string {
	var infos []string
	for _, info := range char.KnownInfos {
		if info.LearnedChapter < chapterID {
			infos = append(infos, fmt.Sprintf("- %s (第%d章得知)", info.Content, info.LearnedChapter))
		}
	}
	return infos
}

// GetCharacterPOV 获取角色视角上下文
func (m *InfoBoundaryManager) GetCharacterPOV(char *model.Character, chapterID int) string {
	knownInfos := m.getKnownInfosBeforeChapter(char, chapterID)

	var pov strings.Builder
	pov.WriteString(fmt.Sprintf("【%s的视角】\n", char.Name))
	pov.WriteString(fmt.Sprintf("状态: %s\n", char.Status))

	if len(knownInfos) > 0 {
		pov.WriteString("已知信息:\n")
		for _, info := range knownInfos {
			pov.WriteString(info + "\n")
		}
	}

	return pov.String()
}

// ExtractInfoFromChapter 从章节提取信息更新
func (m *InfoBoundaryManager) ExtractInfoFromChapter(ctx context.Context, bookName string, chapterID int) (map[string][]model.KnownInfo, error) {
	content, err := m.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	characters, err := m.store.LoadCharacters(bookName)
	if err != nil {
		return nil, err
	}

	// 构建提取提示词
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.Name
	}

	prompt := fmt.Sprintf(`分析以下章节内容，提取每个角色新获得的信息：

【角色列表】
%s

【章节内容】
%s

请按以下JSON格式输出每个角色新获得的信息：
{
  "角色名": [
    {"info_key": "信息标识", "content": "信息内容", "source": "witnessed/hearsay/deduced"}
  ]
}`, strings.Join(charNames, ", "), content)

	result, err := m.llmClient.Call(ctx, prompt, "auditor")
	if err != nil {
		return nil, err
	}

	// 解析结果（简化处理）
	infoMap := make(map[string][]model.KnownInfo)

	for _, char := range characters {
		if strings.Contains(result, char.Name) {
			infoMap[char.Name] = append(infoMap[char.Name], model.KnownInfo{
				ID:             fmt.Sprintf("info_%d_%s", chapterID, char.Name),
				InfoKey:        "extracted",
				Content:        "从章节中获取的信息",
				LearnedChapter: chapterID,
				Source:         model.SourceWitnessed,
			})
		}
	}

	return infoMap, nil
}

// ExtractAndSave 提取信息并持久化到角色数据
func (m *InfoBoundaryManager) ExtractAndSave(ctx context.Context, bookName string, chapterID int) (map[string][]model.KnownInfo, error) {
	infoMap, err := m.ExtractInfoFromChapter(ctx, bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if len(infoMap) == 0 {
		return infoMap, nil
	}

	// 持久化到角色数据
	characters, err := m.store.LoadCharacters(bookName)
	if err != nil {
		return infoMap, err
	}

	for _, char := range characters {
		if infos, ok := infoMap[char.Name]; ok {
			char.KnownInfos = append(char.KnownInfos, infos...)
		}
	}

	if err := m.store.SaveCharacters(bookName, characters); err != nil {
		return infoMap, err
	}

	return infoMap, nil
}