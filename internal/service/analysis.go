package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"ai-writer/internal/llm"
	"ai-writer/internal/store"
)

// AnalysisService 拆书分析服务
type AnalysisService struct {
	llmClient llm.Client
	store     *store.JSONStore
}

// NewAnalysisService 创建拆书分析服务
func NewAnalysisService(llmClient llm.Client, jsonStore *store.JSONStore) *AnalysisService {
	return &AnalysisService{
		llmClient: llmClient,
		store:     jsonStore,
	}
}

// ParsedChapter 解析后的章节
type ParsedChapter struct {
	Num       int    `json:"num"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	WordCount int    `json:"word_count"`
}

// ParseResult 解析结果
type ParseResult struct {
	Success      bool            `json:"success"`
	Message      string          `json:"message,omitempty"`
	Chapters     []ParsedChapter `json:"chapters"`
	TotalWords   int             `json:"total_words"`
	ChapterCount int             `json:"chapter_count"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Characters   []CharacterAnalysis `json:"characters"`
	PlotPoints   []PlotPoint         `json:"plot_points"`
	WorldSetting string              `json:"world_setting"`
	WritingStyle string              `json:"writing_style"`
	Summary      string              `json:"summary"`
}

// CharacterAnalysis 人物分析
type CharacterAnalysis struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Traits      []string `json:"traits"`
	FirstAppear int      `json:"first_appear"`
	Description string   `json:"description"`
}

// PlotPoint 剧情点
type PlotPoint struct {
	Chapter     int    `json:"chapter"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"` // conflict/climax/turning_point
}

// ParseTXT 解析TXT文件内容
func (s *AnalysisService) ParseTXT(content string) *ParseResult {
	// 章节匹配模式
	patterns := []string{
		`(?m)^第[一二三四五六七八九十百千万零0-9]+[章节回卷部篇].*?(?=第[一二三四五六七八九十百千万零0-9]+[章节回卷部篇]|$)`,
		`(?m)^第\s*[0-9]+\s*[章节回卷部篇]?.*?(?=第\s*[0-9]+\s*[章节回卷部篇]?|$)`,
		`(?m)^Chapter\s+\d+.*?(?=^Chapter\s+\d+|$)`,
	}

	var chapters []ParsedChapter

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(content, -1)

		if len(matches) >= 3 {
			for i, match := range matches {
				lines := strings.Split(strings.TrimSpace(match), "\n")
				title := strings.TrimSpace(lines[0])
				if len(title) > 100 {
					title = title[:100]
				}

				text := ""
				if len(lines) > 1 {
					text = strings.Join(lines[1:], "\n")
				} else {
					text = match
				}

				// 提取章节号
				num := extractChapterNum(title)
				if num == 0 {
					num = i + 1
				}

				chapters = append(chapters, ParsedChapter{
					Num:       num,
					Title:     title,
					Content:   strings.TrimSpace(text),
					WordCount: countChineseChars(text),
				})
			}
			break
		}
	}

	// 如果没有找到章节，将整个内容作为一章
	if len(chapters) == 0 {
		chapters = []ParsedChapter{
			{
				Num:       1,
				Title:     "全文",
				Content:   content,
				WordCount: countChineseChars(content),
			},
		}
	}

	totalWords := 0
	for _, ch := range chapters {
		totalWords += ch.WordCount
	}

	return &ParseResult{
		Success:      true,
		Chapters:     chapters,
		TotalWords:   totalWords,
		ChapterCount: len(chapters),
	}
}

// AnalyzeChapters 分析章节内容
func (s *AnalysisService) AnalyzeChapters(ctx context.Context, chapters []ParsedChapter, analysisType string) (*AnalysisResult, error) {
	// 取前几章作为样本
	sampleSize := 5
	if len(chapters) < sampleSize {
		sampleSize = len(chapters)
	}

	var sampleContent strings.Builder
	for i := 0; i < sampleSize; i++ {
		sampleContent.WriteString(fmt.Sprintf("\n【第%d章: %s】\n", chapters[i].Num, chapters[i].Title))
		if len(chapters[i].Content) > 2000 {
			sampleContent.WriteString(chapters[i].Content[:2000])
		} else {
			sampleContent.WriteString(chapters[i].Content)
		}
	}

	prompt := fmt.Sprintf(`请分析以下小说片段，提取关键信息。

小说内容:
%s

请提取并分析：
1. 主要人物（姓名、角色定位、性格特点）
2. 主要剧情点
3. 世界观设定
4. 写作风格特点
5. 内容摘要

请用JSON格式输出：
{
  "characters": [
    {"name": "姓名", "role": "角色", "traits": ["特点"], "first_appear": 1, "description": "描述"}
  ],
  "plot_points": [
    {"chapter": 1, "title": "标题", "description": "描述", "type": "类型"}
  ],
  "world_setting": "世界观设定描述",
  "writing_style": "写作风格特点",
  "summary": "内容摘要"
}`, sampleContent.String())

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var analysis AnalysisResult
	if err := parseJSON(result, &analysis); err != nil {
		analysis.Summary = extractJSONValue(result, "summary")
		analysis.WorldSetting = extractJSONValue(result, "world_setting")
		analysis.WritingStyle = extractJSONValue(result, "writing_style")
	}

	return &analysis, nil
}

// ExtractOutline 提取大纲
func (s *AnalysisService) ExtractOutline(ctx context.Context, chapters []ParsedChapter) ([]*TreeNode, error) {
	var chapterInfos []string
	for _, ch := range chapters {
		chapterInfos = append(chapterInfos, fmt.Sprintf("第%d章: %s", ch.Num, ch.Title))
	}

	prompt := fmt.Sprintf(`请根据以下章节标题，提取小说的整体大纲结构。

章节列表:
%s

请输出分卷结构和大纲：
{
  "volumes": [
    {
      "id": "vol_1",
      "label": "第一卷标题",
      "outline": "本卷主要内容",
      "children": [
        {"id": "chap_1", "label": "章节标题", "outline": "章节概要"}
      ]
    }
  ]
}`, strings.Join(chapterInfos, "\n"))

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var outline struct {
		Volumes []*TreeNode `json:"volumes"`
	}
	if err := parseJSON(result, &outline); err != nil {
		return nil, err
	}

	return outline.Volumes, nil
}

// CompareWithMyWork 与自己的作品对比
func (s *AnalysisService) CompareWithMyWork(ctx context.Context, analysis *AnalysisResult, myBookName string) (string, error) {
	// 加载用户作品信息
	myChars, _ := s.store.LoadCharacters(myBookName)
	myWorldView, _ := s.store.LoadWorldView(myBookName)

	var myCharNames []string
	for _, c := range myChars {
		myCharNames = append(myCharNames, c.Name)
	}

	prompt := fmt.Sprintf(`请对比分析两部作品的人物设定和世界观。

参考作品人物: %v
参考作品世界观: %s

我的作品人物: %v
我的作品世界观: %s

请分析：
1. 人物设定的相似与差异
2. 世界观的相似与差异
3. 可借鉴的创作手法
4. 避免重复的建议

请输出详细分析。`,
		getCharNames(analysis.Characters),
		analysis.WorldSetting,
		myCharNames,
		myWorldView)

	return s.llmClient.Call(ctx, prompt, "architect")
}

func getCharNames(chars []CharacterAnalysis) []string {
	var names []string
	for _, c := range chars {
		names = append(names, c.Name)
	}
	return names
}

// extractChapterNum 提取章节号
func extractChapterNum(title string) int {
	// 匹配数字
	re := regexp.MustCompile(`[0-9]+`)
	if match := re.FindString(title); match != "" {
		var num int
		fmt.Sscanf(match, "%d", &num)
		return num
	}

	// 匹配中文数字
	chineseNums := map[rune]int{
		'一': 1, '二': 2, '三': 3, '四': 4, '五': 5,
		'六': 6, '七': 7, '八': 8, '九': 9, '十': 10,
		'百': 100, '千': 1000,
	}

	result := 0
	for _, r := range title {
		if n, ok := chineseNums[r]; ok {
			if n == 10 {
				if result == 0 {
					result = 10
				} else {
					result *= 10
				}
			} else if result >= 10 {
				result += n
			} else {
				result = n
			}
		}
	}

	return result
}

func countChineseChars(s string) int {
	count := 0
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			count++
		}
	}
	return count
}