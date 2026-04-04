package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/service"
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量生成流水线",
	Long: `批量生成章节内容的流水线工具。

支持断点续传、进度追踪、失败重试。

示例:
  # 批量生成第1-50章
  ai-writer batch generate --from 1 --to 50

  # 从上次中断处继续
  ai-writer batch continue

  # 查看进度
  ai-writer batch status

  # 重置进度
  ai-writer batch reset`,
}

var batchGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "批量生成章节",
	Long: `批量生成指定范围的章节。

生成时会自动：
1. 构建完整写作上下文
2. 保存生成内容
3. 记录进度

支持 Ctrl+C 中断，下次可用 continue 继续。`,
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		from, _ := cmd.Flags().GetInt("from")
		to, _ := cmd.Flags().GetInt("to")
		stream, _ := cmd.Flags().GetBool("stream")
		retry, _ := cmd.Flags().GetInt("retry")

		if from < 1 || to < from {
			fmt.Fprintf(os.Stderr, "错误: 无效的章节范围 %d-%d\n", from, to)
			return
		}

		// 初始化 LLM 客户端
		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 创建写作服务
		writerService := service.NewWriterService(llmClient, jsonStore, &cfg.Prompts)

		// 加载进度
		progress := loadProgress(bookName)
		if progress == nil {
			progress = &BatchProgress{
				BookName:  bookName,
				From:      from,
				To:        to,
				Current:   from,
				Completed: []int{},
				Failed:    []int{},
				StartedAt: time.Now(),
			}
		}

		// 检查范围是否匹配
		if progress.From != from || progress.To != to {
			// 新任务，重置进度
			progress = &BatchProgress{
				BookName:  bookName,
				From:      from,
				To:        to,
				Current:   from,
				Completed: []int{},
				Failed:    []int{},
				StartedAt: time.Now(),
			}
		}

		fmt.Printf("🚀 批量生成: 第%d章 到 第%d章\n", from, to)
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("  进度: %d/%d\n", len(progress.Completed), to-from+1)
		fmt.Printf("  当前: 第%d章\n", progress.Current)
		if len(progress.Failed) > 0 {
			fmt.Printf("  失败: %v\n", progress.Failed)
		}
		fmt.Println("────────────────────────────────")
		fmt.Println("按 Ctrl+C 可中断，下次用 continue 继续")
		fmt.Println()

		// 设置信号处理
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 生成循环
		ctx := context.Background()

		for chapterID := progress.Current; chapterID <= to; chapterID++ {
			// 检查是否已完成
			if contains(progress.Completed, chapterID) {
				continue
			}

			select {
			case <-sigChan:
				// 保存进度并退出
				progress.Current = chapterID
				progress.UpdatedAt = time.Now()
				saveProgress(bookName, progress)
				fmt.Println("\n\n⏸️  已中断，进度已保存")
				fmt.Printf("使用 'ai-writer batch continue' 继续\n")
				return
			default:
			}

			fmt.Printf("\n📖 生成第%d章...\n", chapterID)

			var content string
			var genErr error

			for attempt := 0; attempt <= retry; attempt++ {
				if attempt > 0 {
					fmt.Printf("  重试 %d/%d...\n", attempt, retry)
					time.Sleep(2 * time.Second)
				}

				genCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)

				if stream {
					content, genErr = streamGenerate(genCtx, writerService, bookName, chapterID)
				} else {
					content, genErr = writerService.WriteChapter(genCtx, bookName, chapterID, "")
				}

				cancel()

				if genErr == nil {
					break
				}
			}

			if genErr != nil {
				fmt.Printf("  ❌ 失败: %v\n", genErr)
				progress.Failed = append(progress.Failed, chapterID)
				continue
			}

			// 保存内容
			if err := jsonStore.SaveChapterContent(bookName, chapterID, content); err != nil {
				fmt.Printf("  ❌ 保存失败: %v\n", err)
				progress.Failed = append(progress.Failed, chapterID)
				continue
			}

			wordCount := countChineseChars(content)
			fmt.Printf("  ✅ 完成: %d 字\n", wordCount)

			progress.Completed = append(progress.Completed, chapterID)
			progress.Current = chapterID + 1
			progress.UpdatedAt = time.Now()

			// 保存进度
			saveProgress(bookName, progress)
		}

		// 完成
		fmt.Println("\n═══════════════════════════════════════")
		fmt.Printf("🎉 批量生成完成！\n")
		fmt.Printf("  成功: %d 章\n", len(progress.Completed))
		if len(progress.Failed) > 0 {
			fmt.Printf("  失败: %d 章 (%v)\n", len(progress.Failed), progress.Failed)
		}

		// 删除进度文件
		deleteProgress(bookName)
	},
}

var batchContinueCmd = &cobra.Command{
	Use:   "continue",
	Short: "继续上次中断的任务",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		progress := loadProgress(bookName)
		if progress == nil {
			fmt.Println("没有未完成的任务")
			return
		}

		// 使用父命令的逻辑继续生成
		fmt.Printf("继续生成: 第%d章 到 第%d章\n", progress.Current, progress.To)

		// 调用生成逻辑
		batchGenerateCmd.Run(cmd, []string{})
	},
}

var batchStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看批量生成进度",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		progress := loadProgress(bookName)
		if progress == nil {
			fmt.Println("没有进行中的批量任务")
			return
		}

		fmt.Println("📊 批量生成进度")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("  书名: %s\n", progress.BookName)
		fmt.Printf("  范围: 第%d章 - 第%d章\n", progress.From, progress.To)
		fmt.Printf("  当前进度: %d/%d (%.1f%%)\n",
			len(progress.Completed), progress.To-progress.From+1,
			float64(len(progress.Completed))/float64(progress.To-progress.From+1)*100)
		fmt.Printf("  当前章节: 第%d章\n", progress.Current)

		if len(progress.Completed) > 0 {
			fmt.Printf("  已完成: %v\n", progress.Completed)
		}

		if len(progress.Failed) > 0 {
			fmt.Printf("  失败: %v\n", progress.Failed)
		}

		fmt.Printf("  开始时间: %s\n", progress.StartedAt.Format("2006-01-02 15:04:05"))
		if !progress.UpdatedAt.IsZero() {
			fmt.Printf("  更新时间: %s\n", progress.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	},
}

var batchResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置批量生成进度",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Print("确定要重置进度吗？此操作不可恢复！输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		deleteProgress(bookName)
		fmt.Println("✅ 进度已重置")
	},
}

// BatchProgress 批量生成进度
type BatchProgress struct {
	BookName  string    `json:"book_name"`
	From      int       `json:"from"`
	To        int       `json:"to"`
	Current   int       `json:"current"`
	Completed []int     `json:"completed"`
	Failed    []int     `json:"failed"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.AddCommand(batchGenerateCmd)
	batchCmd.AddCommand(batchContinueCmd)
	batchCmd.AddCommand(batchStatusCmd)
	batchCmd.AddCommand(batchResetCmd)

	batchGenerateCmd.Flags().Int("from", 1, "起始章节")
	batchGenerateCmd.Flags().Int("to", 10, "结束章节")
	batchGenerateCmd.Flags().Bool("stream", true, "流式输出")
	batchGenerateCmd.Flags().Int("retry", 2, "失败重试次数")

	batchResetCmd.Flags().BoolP("force", "f", false, "强制重置，不确认")
}

// loadProgress 加载进度
func loadProgress(bookName string) *BatchProgress {
	data, err := os.ReadFile(getProgressFile(bookName))
	if err != nil {
		return nil
	}

	var progress BatchProgress
	if err := parseJSONFromBytes(data, &progress); err != nil {
		return nil
	}

	return &progress
}

// saveProgress 保存进度
func saveProgress(bookName string, progress *BatchProgress) {
	data, _ := printJSONToBytes(progress)
	os.WriteFile(getProgressFile(bookName), data, 0644)
}

// deleteProgress 删除进度
func deleteProgress(bookName string) {
	os.Remove(getProgressFile(bookName))
}

// getProgressFile 获取进度文件路径
func getProgressFile(bookName string) string {
	return fmt.Sprintf("data/projects/%s/.progress.json", bookName)
}

// streamGenerate 流式生成
func streamGenerate(ctx context.Context, writerService *service.WriterService, bookName string, chapterID int) (string, error) {
	stream, err := writerService.WriteChapterStream(ctx, bookName, chapterID, "")
	if err != nil {
		return "", err
	}

	var content strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			return "", chunk.Error
		}
		if chunk.Done {
			break
		}
		content.WriteString(chunk.Content)
	}

	return content.String(), nil
}

// contains 检查切片是否包含元素
func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// parseJSONFromBytes 解析 JSON
func parseJSONFromBytes(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// printJSONToBytes 序列化 JSON
func printJSONToBytes(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}