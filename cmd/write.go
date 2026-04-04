package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/llm"
	"ai-writer/internal/service"
)

var (
	writeStream  bool
	writeWords   int
	writeOutline string
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write <章节号>",
	Short: "AI 生成章节",
	Long: `使用 AI 生成章节内容。

示例:
  ai-writer write 1                          # 生成第1章
  ai-writer write 1 --stream                 # 流式生成
  ai-writer write 1 --outline "主角觉醒"      # 指定大纲
  ai-writer write next                       # 写下一章`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])

		// 初始化 LLM 客户端
		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 创建写作服务
		writerService := service.NewWriterService(llmClient, jsonStore, &cfg.Prompts)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if writeStream {
			// 流式生成
			streamWrite(ctx, writerService, bookName, chapterID, writeOutline)
		} else {
			// 普通生成
			normalWrite(ctx, writerService, bookName, chapterID, writeOutline)
		}
	},
}

// continueCmd represents the continue command
var continueCmd = &cobra.Command{
	Use:   "continue <章节号>",
	Short: "续写章节",
	Long: `续写现有章节内容。

示例:
  ai-writer continue 1              # 续写第1章
  ai-writer continue 1 --words 500  # 续写500字`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])

		// 加载现有内容
		existingContent, err := jsonStore.LoadChapterContent(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if existingContent == "" {
			fmt.Fprintf(os.Stderr, "错误: 章节 %d 没有内容，请使用 write 命令生成\n", chapterID)
			return
		}

		// 初始化 LLM 客户端
		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		writerService := service.NewWriterService(llmClient, jsonStore, &cfg.Prompts)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		fmt.Printf("正在续写第%d章...\n", chapterID)

		content, err := writerService.ContinueChapter(ctx, bookName, chapterID, existingContent, writeWords)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 合并内容
		newContent := existingContent + "\n\n" + content

		// 保存
		if err := jsonStore.SaveChapterContent(bookName, chapterID, newContent); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已续写 %d 字\n", countChineseChars(content))
		if verbose {
			fmt.Println("────────────────────────────────")
			fmt.Println(content)
		}
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(continueCmd)

	// write 命令选项
	writeCmd.Flags().BoolVarP(&writeStream, "stream", "s", false, "流式输出")
	writeCmd.Flags().IntVarP(&writeWords, "words", "w", 3000, "目标字数")
	writeCmd.Flags().StringVarP(&writeOutline, "outline", "O", "", "章节大纲")

	// continue 命令选项
	continueCmd.Flags().IntVarP(&writeWords, "words", "w", 500, "续写字数")
}

// initLLMClient 初始化 LLM 客户端
func initLLMClient() (llm.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	llmConfig := &llm.Config{
		Provider:     cfg.LLM.Provider,
		APIKey:       cfg.LLM.APIKey,
		BaseURL:      cfg.LLM.BaseURL,
		Models:       cfg.LLM.Models,
		Temperatures: cfg.LLM.Temperatures,
		MaxRetries:   cfg.LLM.MaxRetries,
		Timeout:      cfg.LLM.Timeout,
	}

	return llm.NewClient(llmConfig), nil
}

// normalWrite 普通生成
func normalWrite(ctx context.Context, writerService *service.WriterService, bookName string, chapterID int, outline string) {
	fmt.Printf("正在生成第%d章...\n", chapterID)

	content, err := writerService.WriteChapter(ctx, bookName, chapterID, outline)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	// 保存
	if err := jsonStore.SaveChapterContent(bookName, chapterID, content); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 已生成 %d 字\n", countChineseChars(content))
	if verbose {
		fmt.Println("────────────────────────────────")
		fmt.Println(content)
	}
}

// streamWrite 流式生成
func streamWrite(ctx context.Context, writerService *service.WriterService, bookName string, chapterID int, outline string) {
	fmt.Printf("正在流式生成第%d章...\n", chapterID)
	fmt.Println("────────────────────────────────")

	stream, err := writerService.WriteChapterStream(ctx, bookName, chapterID, outline)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	var content strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			fmt.Fprintf(os.Stderr, "\n错误: %v\n", chunk.Error)
			return
		}

		if chunk.Done {
			break
		}

		fmt.Print(chunk.Content)
		content.WriteString(chunk.Content)
	}

	fmt.Println("\n────────────────────────────────")

	// 保存
	fullContent := content.String()
	if err := jsonStore.SaveChapterContent(bookName, chapterID, fullContent); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 已生成 %d 字\n", countChineseChars(fullContent))
}

// parseChapterRange 解析章节范围
func parseChapterRange(rangeStr string, bookName string) []int {
	chapters, _ := jsonStore.LoadChapters(bookName)
	if len(chapters) == 0 {
		return []int{}
	}

	// all
	if rangeStr == "all" {
		ids := make([]int, len(chapters))
		for i, ch := range chapters {
			ids[i] = ch.ID
		}
		return ids
	}

	// 单章节
	if !strings.Contains(rangeStr, "-") {
		return []int{parseChapterID(rangeStr)}
	}

	// 范围
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return []int{parseChapterID(rangeStr)}
	}

	start := parseChapterID(parts[0])
	end := parseChapterID(parts[1])

	var ids []int
	for i := start; i <= end; i++ {
		ids = append(ids, i)
	}
	return ids
}

// 确保 io 包被使用
var _ = io.EOF