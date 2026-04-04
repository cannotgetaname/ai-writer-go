package service

import (
	"context"
	"fmt"
	"strings"

	"ai-writer/internal/config"
	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/store"
)

// WriterService AI 写作服务
type WriterService struct {
	llmClient      llm.Client
	store          *store.JSONStore
	prompts        *config.PromptsConfig
	contextService *ContextService
}

// NewWriterService 创建写作服务
func NewWriterService(llmClient llm.Client, store *store.JSONStore, prompts *config.PromptsConfig) *WriterService {
	return &WriterService{
		llmClient:      llmClient,
		store:          store,
		prompts:        prompts,
		contextService: NewContextService(store),
	}
}

// WriteChapter 生成章节内容
func (s *WriterService) WriteChapter(ctx context.Context, bookName string, chapterID int, outline string) (string, error) {
	// 加载上下文
	context, err := s.buildWritingContext(bookName, chapterID)
	if err != nil {
		return "", err
	}

	// 加载章节信息
	chapters, err := s.store.LoadChapters(bookName)
	if err != nil {
		return "", err
	}

	var chapter *model.Chapter
	for _, ch := range chapters {
		if ch.ID == chapterID {
			chapter = ch
			break
		}
	}

	if chapter == nil {
		return "", fmt.Errorf("章节 %d 不存在", chapterID)
	}

	// 构建提示词
	userPrompt := s.buildUserPrompt(context, chapter, outline)

	// 调用 LLM
	systemPrompt := s.prompts.WriterSystem
	content, err := s.llmClient.CallWithSystem(ctx, systemPrompt, userPrompt, "writer")
	if err != nil {
		return "", err
	}

	return content, nil
}

// WriteChapterStream 流式生成章节
func (s *WriterService) WriteChapterStream(ctx context.Context, bookName string, chapterID int, outline string) (<-chan llm.StreamChunk, error) {
	// 加载上下文
	context, err := s.buildWritingContext(bookName, chapterID)
	if err != nil {
		return nil, err
	}

	// 加载章节信息
	chapters, err := s.store.LoadChapters(bookName)
	if err != nil {
		return nil, err
	}

	var chapter *model.Chapter
	for _, ch := range chapters {
		if ch.ID == chapterID {
			chapter = ch
			break
		}
	}

	if chapter == nil {
		return nil, fmt.Errorf("章节 %d 不存在", chapterID)
	}

	// 构建提示词
	userPrompt := s.buildUserPrompt(context, chapter, outline)

	// 流式调用 LLM
	systemPrompt := s.prompts.WriterSystem
	return s.llmClient.StreamWithSystem(ctx, systemPrompt, userPrompt, "writer")
}

// ContinueChapter 续写章节
func (s *WriterService) ContinueChapter(ctx context.Context, bookName string, chapterID int, existingContent string, words int) (string, error) {
	// 加载上下文
	context, err := s.buildWritingContext(bookName, chapterID)
	if err != nil {
		return "", err
	}

	// 构建续写提示词
	userPrompt := fmt.Sprintf(`请续写以下内容，续写约%d字：

【已有内容】
%s

【上下文信息】
%s

请自然衔接上文，保持风格一致。`, words, existingContent, context)

	systemPrompt := s.prompts.WriterSystem
	content, err := s.llmClient.CallWithSystem(ctx, systemPrompt, userPrompt, "writer")
	if err != nil {
		return "", err
	}

	return content, nil
}

// RewriteChapter 重写章节
func (s *WriterService) RewriteChapter(ctx context.Context, bookName string, chapterID int, instruction string) (string, error) {
	// 加载现有内容
	content, err := s.store.LoadChapterContent(bookName, chapterID)
	if err != nil {
		return "", err
	}

	if content == "" {
		return "", fmt.Errorf("章节 %d 没有内容", chapterID)
	}

	// 构建重写提示词
	userPrompt := fmt.Sprintf(`请根据以下要求重写章节内容：

【原内容】
%s

【重写要求】
%s

请保持故事核心不变，按要求进行修改。`, content, instruction)

	systemPrompt := s.prompts.WriterSystem
	newContent, err := s.llmClient.CallWithSystem(ctx, systemPrompt, userPrompt, "writer")
	if err != nil {
		return "", err
	}

	return newContent, nil
}

// buildWritingContext 构建写作上下文
func (s *WriterService) buildWritingContext(bookName string, chapterID int) (string, error) {
	// 使用增强的上下文服务
	ctx, err := s.contextService.BuildContext(bookName, chapterID, 2000)
	if err != nil {
		return "", err
	}

	return s.contextService.FormatContext(ctx), nil
}

// buildUserPrompt 构建用户提示词
func (s *WriterService) buildUserPrompt(context string, chapter *model.Chapter, customOutline string) string {
	var prompt strings.Builder

	prompt.WriteString("请根据以下信息撰写章节内容：\n\n")

	if context != "" {
		prompt.WriteString(context)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString(fmt.Sprintf("【章节信息】\n"))
	prompt.WriteString(fmt.Sprintf("第%d章: %s\n", chapter.ID, chapter.Title))

	if customOutline != "" {
		prompt.WriteString(fmt.Sprintf("大纲: %s\n", customOutline))
	} else if chapter.Outline != "" {
		prompt.WriteString(fmt.Sprintf("大纲: %s\n", chapter.Outline))
	}

	prompt.WriteString("\n请撰写完整的章节内容，字数在3000-5000字之间。")

	return prompt.String()
}

// RewriteParagraph 重写段落
func (s *WriterService) RewriteParagraph(ctx context.Context, bookName string, chapterID int, paragraphID string, rewritePrompt string) (string, error) {
	// 加载段落
	paragraphs, err := s.store.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		return "", err
	}

	// 查找目标段落
	var targetParagraph *model.Paragraph
	for _, p := range paragraphs.Paragraphs {
		if p.ID == paragraphID {
			targetParagraph = &p
			break
		}
	}

	if targetParagraph == nil {
		return "", fmt.Errorf("段落 %s 不存在", paragraphID)
	}

	// 获取上下文
	context := getParagraphContext(paragraphID, paragraphs)

	// 构建重写提示词
	userPrompt := fmt.Sprintf(`请根据以下要求重写段落：

【原文】
%s

【重写要求】
%s

【上下文】
%s

请保持风格一致，自然衔接上下文。`, targetParagraph.Text, rewritePrompt, context)

	systemPrompt := s.prompts.WriterSystem
	newContent, err := s.llmClient.CallWithSystem(ctx, systemPrompt, userPrompt, "writer")
	if err != nil {
		return "", err
	}

	return newContent, nil
}

// getParagraphContext 获取段落上下文
func getParagraphContext(paragraphID string, paragraphs *model.ChapterParagraphs) string {
	var context strings.Builder

	for i, p := range paragraphs.Paragraphs {
		if p.ID == paragraphID {
			// 前一段
			if i > 0 {
				context.WriteString("前一段:\n")
				if len(paragraphs.Paragraphs[i-1].Text) > 100 {
					context.WriteString(paragraphs.Paragraphs[i-1].Text[:100] + "...")
				} else {
					context.WriteString(paragraphs.Paragraphs[i-1].Text)
				}
				context.WriteString("\n\n")
			}
			// 后一段
			if i < len(paragraphs.Paragraphs)-1 {
				context.WriteString("后一段:\n")
				if len(paragraphs.Paragraphs[i+1].Text) > 100 {
					context.WriteString(paragraphs.Paragraphs[i+1].Text[:100] + "...")
				} else {
					context.WriteString(paragraphs.Paragraphs[i+1].Text)
				}
			}
			break
		}
	}

	return context.String()
}