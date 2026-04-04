package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/llm"
	"ai-writer/internal/model"
	"ai-writer/internal/service"
)

// reviewCmd represents the review command
var reviewCmd = &cobra.Command{
	Use:   "review <章节号>",
	Short: "AI 审稿",
	Long: `使用 AI 审稿章节内容，标记问题段落并给出修改建议。

审稿维度:
1. 人设一致性 - 人物言行是否符合其性格
2. 剧情逻辑 - 是否有前后矛盾
3. 爽点节奏 - 是否拖沓，是否有期待感
4. 文笔表达 - 描写是否生动

示例:
  ai-writer review 1              # 审稿第1章
  ai-writer review 1 --fix        # 审稿并进入修复模式
  ai-writer review 1-10           # 审稿第1到10章`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 初始化 LLM 客户端
		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		reviewService := service.NewReviewService(llmClient, jsonStore, &cfg.Prompts)

		// 解析章节范围
		chapterIDs := parseChapterRange(args[0], bookName)

		fixMode, _ := cmd.Flags().GetBool("fix")

		for _, chapterID := range chapterIDs {
			fmt.Printf("\n📄 正在审稿第%d章...\n", chapterID)
			fmt.Println("────────────────────────────────")

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

			result, err := reviewService.ReviewChapter(ctx, bookName, chapterID)
			cancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				continue
			}

			// 显示审稿报告
			displayReviewResult(result)

			// 修复模式
			if fixMode && len(result.Issues) > 0 {
				fixModeInteractive(bookName, chapterID, result, llmClient)
			}
		}
	},
}

// rewriteCmd represents the rewrite command
var rewriteCmd = &cobra.Command{
	Use:   "rewrite <章节号>",
	Short: "重写段落",
	Long: `根据审稿意见或自定义指令重写段落。

示例:
  ai-writer rewrite 1 --paragraph 3                     # 根据审稿意见重写段落3
  ai-writer rewrite 1 --paragraph 3 --instruction "增加战斗细节"  # 自定义指令重写`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		paragraphID, _ := cmd.Flags().GetString("paragraph")
		instruction, _ := cmd.Flags().GetString("instruction")

		if paragraphID == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定段落ID (--paragraph)\n")
			return
		}

		// 加载段落
		paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找段落
		var targetParagraph *model.Paragraph
		for _, p := range paragraphs.Paragraphs {
			if p.ID == paragraphID || truncate(p.ID, 8) == paragraphID {
				targetParagraph = &p
				break
			}
		}

		if targetParagraph == nil {
			fmt.Fprintf(os.Stderr, "错误: 段落 %s 不存在\n", paragraphID)
			return
		}

		// 初始化 LLM 客户端
		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		writerService := service.NewWriterService(llmClient, jsonStore, &cfg.Prompts)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		fmt.Printf("正在重写段落 %s...\n", truncate(paragraphID, 8))
		if instruction != "" {
			fmt.Printf("指令: %s\n", instruction)
		}

		// 构建重写提示词
		rewritePrompt := buildRewritePrompt(targetParagraph, instruction, paragraphs)

		newContent, err := writerService.RewriteParagraph(ctx, bookName, chapterID, targetParagraph.ID, rewritePrompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 显示结果
		fmt.Println("\n原文:")
		fmt.Println("────────────────────────────────")
		fmt.Println(truncate(targetParagraph.Text, 200))
		if len(targetParagraph.Text) > 200 {
			fmt.Println("...")
		}

		fmt.Println("\n新文:")
		fmt.Println("────────────────────────────────")
		fmt.Println(truncate(newContent, 200))
		if len(newContent) > 200 {
			fmt.Println("...")
		}

		// 确认是否应用
		fmt.Print("\n是否应用更改？[y/N]: ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "y" || confirm == "Y" {
			if err := jsonStore.UpdateParagraph(bookName, chapterID, targetParagraph.ID, newContent); err != nil {
				fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
				return
			}
			fmt.Println("✅ 已应用更改")
		} else {
			fmt.Println("已取消")
		}
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(rewriteCmd)

	reviewCmd.Flags().Bool("fix", false, "审稿后进入修复模式")

	rewriteCmd.Flags().String("paragraph", "", "段落ID")
	rewriteCmd.Flags().String("instruction", "", "重写指令")
}

// displayReviewResult 显示审稿结果
func displayReviewResult(result *service.ReviewResult) {
	fmt.Printf("📊 综合评分: %d/100\n\n", result.OverallScore)

	if len(result.Issues) > 0 {
		fmt.Println("问题列表:")
		fmt.Println("────────────────────────────────")
		for i, issue := range result.Issues {
			fmt.Printf("\n[%d] [%s][%s] %s\n", i+1, issue.Type, issue.Severity, issue.Description)
			if issue.Location != "" {
				fmt.Printf("    位置: %s\n", issue.Location)
			}
			fmt.Printf("    建议: %s\n", issue.Suggestion)
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Println("\n修改建议:")
		fmt.Println("────────────────────────────────")
		for _, s := range result.Suggestions {
			fmt.Printf("  - %s\n", s)
		}
	}
}

// fixModeInteractive 修复模式交互
func fixModeInteractive(bookName string, chapterID int, result *service.ReviewResult, llmClient llm.Client) {
	fmt.Println("\n────────────────────────────────")
	fmt.Println("进入修复模式")
	fmt.Println("────────────────────────────────")

	// 加载段落
	paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法加载段落: %v\n", err)
		return
	}

	writerService := service.NewWriterService(llmClient, jsonStore, &cfg.Prompts)

	for i, issue := range result.Issues {
		fmt.Printf("\n修复问题 [%d]: %s\n", i+1, issue.Description)
		fmt.Printf("建议: %s\n", issue.Suggestion)
		fmt.Print("是否重写相关段落？[y/n/s(跳过)]: ")

		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "n" || confirm == "s" {
			continue
		}

		if confirm != "y" {
			continue
		}

		// 尝试定位问题段落
		paragraphID := locateIssueParagraph(issue, paragraphs)
		if paragraphID == "" {
			fmt.Println("无法定位问题段落，跳过")
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

		rewritePrompt := fmt.Sprintf("问题: %s\n建议: %s", issue.Description, issue.Suggestion)
		newContent, err := writerService.RewriteParagraph(ctx, bookName, chapterID, paragraphID, rewritePrompt)

		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "重写失败: %v\n", err)
			continue
		}

		fmt.Printf("\n新内容:\n%s\n", truncate(newContent, 300))
		fmt.Print("应用更改？[y/N]: ")

		var apply string
		fmt.Scanln(&apply)

		if apply == "y" || apply == "Y" {
			jsonStore.UpdateParagraph(bookName, chapterID, paragraphID, newContent)
			fmt.Println("✅ 已应用")
		}
	}
}

// locateIssueParagraph 尝试定位问题段落
func locateIssueParagraph(issue service.ReviewIssue, paragraphs *model.ChapterParagraphs) string {
	// 如果问题中包含段落ID
	if issue.Location != "" {
		for _, p := range paragraphs.Paragraphs {
			if p.ID == issue.Location || truncate(p.ID, 8) == issue.Location {
				return p.ID
			}
		}
	}

	// 否则返回第一个段落（简单处理）
	if len(paragraphs.Paragraphs) > 0 {
		return paragraphs.Paragraphs[0].ID
	}

	return ""
}

// buildRewritePrompt 构建重写提示词
func buildRewritePrompt(paragraph *model.Paragraph, instruction string, paragraphs *model.ChapterParagraphs) string {
	var prompt strings.Builder

	// 获取前后段落作为上下文
	context := getParagraphContext(paragraph.ID, paragraphs)

	prompt.WriteString(fmt.Sprintf("请根据以下要求重写段落：\n\n"))
	prompt.WriteString("【原文】\n")
	prompt.WriteString(paragraph.Text)
	prompt.WriteString("\n\n")

	if instruction != "" {
		prompt.WriteString("【重写要求】\n")
		prompt.WriteString(instruction)
		prompt.WriteString("\n\n")
	}

	if context != "" {
		prompt.WriteString("【上下文】\n")
		prompt.WriteString(context)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("请保持风格一致，自然衔接上下文。")

	return prompt.String()
}

// getParagraphContext 获取段落上下文
func getParagraphContext(paragraphID string, paragraphs *model.ChapterParagraphs) string {
	var context strings.Builder

	for i, p := range paragraphs.Paragraphs {
		if p.ID == paragraphID {
			// 前一段
			if i > 0 {
				context.WriteString("前一段:\n")
				context.WriteString(truncate(paragraphs.Paragraphs[i-1].Text, 100))
				context.WriteString("\n\n")
			}
			// 后一段
			if i < len(paragraphs.Paragraphs)-1 {
				context.WriteString("后一段:\n")
				context.WriteString(truncate(paragraphs.Paragraphs[i+1].Text, 100))
			}
			break
		}
	}

	return context.String()
}

// truncate 辅助函数在 root.go 中定义