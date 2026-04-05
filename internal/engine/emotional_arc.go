package engine

import (
	"context"
	"fmt"
	"strings"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// EmotionalArcTracker 情感弧线追踪器
type EmotionalArcTracker struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewEmotionalArcTracker 创建情感弧线追踪器
func NewEmotionalArcTracker(llmClient llm.Client, store *store.JSONStore) *EmotionalArcTracker {
	return &EmotionalArcTracker{
		llmClient: llmClient,
		store:     store,
	}
}

// TrackEmotion 追踪章节中的情感变化
func (t *EmotionalArcTracker) TrackEmotion(ctx context.Context, bookName string, chapterID int) (map[string]model.EmotionPoint, error) {
	// 加载章节内容
	content, err := t.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	// 加载人物
	characters, err := t.store.LoadCharacters(bookName)
	if err != nil {
		return nil, err
	}

	emotionMap := make(map[string]model.EmotionPoint)

	for _, char := range characters {
		// 只追踪主要角色
		if char.Role != "主角" && char.Role != "配角" {
			continue
		}

		// 构建提示词
		prompt := fmt.Sprintf(`分析以下章节中角色"%s"的情感状态：

【章节内容】
%s

请分析该角色的情感：
1. 主要情绪类型（愤怒/悲伤/喜悦/恐惧/惊讶/厌恶等）
2. 情绪强度（1-10）
3. 触发事件

请用JSON格式输出：
{"emotion": "情绪类型", "intensity": 强度数字, "trigger": "触发事件"}`, char.Name, content)

		result, err := t.llmClient.Call(ctx, prompt, "auditor")
		if err != nil {
			continue
		}

		// 解析结果
		point := model.EmotionPoint{
			ChapterID: chapterID,
		}

		if strings.Contains(result, `"emotion"`) {
			point.Emotion = extractJSONValue(result, "emotion")
			point.Trigger = extractJSONValue(result, "trigger")
			intensityStr := extractJSONValue(result, "intensity")
			if intensityStr != "" {
				fmt.Sscanf(intensityStr, "%d", &point.Intensity)
			}
			if point.Intensity == 0 {
				point.Intensity = 5
			}
		}

		emotionMap[char.Name] = point
	}

	return emotionMap, nil
}

// GetArcData 获取情感弧线数据
func (t *EmotionalArcTracker) GetArcData(bookName string, charName string) ([]model.EmotionPoint, error) {
	characters, err := t.store.LoadCharacters(bookName)
	if err != nil {
		return nil, err
	}

	for _, char := range characters {
		if char.Name == charName {
			return char.EmotionalArc, nil
		}
	}

	return nil, fmt.Errorf("角色 %s 不存在", charName)
}

// DetectArcComplete 检测角色弧线是否完成
func (t *EmotionalArcTracker) DetectArcComplete(bookName string, charName string) (bool, string) {
	arc, err := t.GetArcData(bookName, charName)
	if err != nil {
		return false, err.Error()
	}

	if len(arc) < 3 {
		return false, "弧线数据不足"
	}

	// 简单判断：检查情感是否趋于稳定
	lastEmotions := arc[len(arc)-3:]
	stable := true
	for i := 1; i < len(lastEmotions); i++ {
		if lastEmotions[i].Emotion != lastEmotions[0].Emotion {
			stable = false
			break
		}
	}

	if stable {
		return true, "角色情感趋于稳定，弧线可能已完成"
	}

	return false, "角色情感仍在变化中"
}

// GetEmotionSummary 获取情感摘要
func (t *EmotionalArcTracker) GetEmotionSummary(bookName string, charName string) string {
	arc, err := t.GetArcData(bookName, charName)
	if err != nil {
		return err.Error()
	}

	if len(arc) == 0 {
		return "暂无情感数据"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("%s 的情感弧线:\n", charName))

	for _, point := range arc {
		summary.WriteString(fmt.Sprintf("  第%d章: %s (强度: %d) - %s\n",
			point.ChapterID, point.Emotion, point.Intensity, point.Trigger))
	}

	// 计算主导情绪
	emotionCount := make(map[string]int)
	for _, point := range arc {
		emotionCount[point.Emotion]++
	}

	var dominantEmotion string
	maxCount := 0
	for emotion, count := range emotionCount {
		if count > maxCount {
			maxCount = count
			dominantEmotion = emotion
		}
	}

	summary.WriteString(fmt.Sprintf("\n主导情绪: %s", dominantEmotion))

	return summary.String()
}

// TrackAndSave 追踪情感并持久化到角色数据
func (t *EmotionalArcTracker) TrackAndSave(ctx context.Context, bookName string, chapterID int) (map[string]model.EmotionPoint, error) {
	emotions, err := t.TrackEmotion(ctx, bookName, chapterID)
	if err != nil {
		return nil, err
	}

	if len(emotions) == 0 {
		return emotions, nil
	}

	// 持久化到角色数据
	characters, err := t.store.LoadCharacters(bookName)
	if err != nil {
		return emotions, err
	}

	for _, char := range characters {
		if point, ok := emotions[char.Name]; ok {
			char.EmotionalArc = append(char.EmotionalArc, point)
		}
	}

	if err := t.store.SaveCharacters(bookName, characters); err != nil {
		return emotions, err
	}

	return emotions, nil
}