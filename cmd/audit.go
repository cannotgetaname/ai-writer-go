package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
	"ai-writer/internal/service"
)

// auditCmd represents the audit command
var auditCmd = &cobra.Command{
	Use:   "audit <章节号>",
	Short: "状态审计章节",
	Long: `使用 AI 分析章节内容，提取状态变更（人物状态、物品归属等）。

示例:
  ai-writer audit 1              # 审计第1章
  ai-writer audit 1-10           # 审计第1到10章`,
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

		// 创建审稿服务
		reviewService := service.NewReviewService(llmClient, jsonStore, &cfg.Prompts)

		// 解析章节范围
		chapterIDs := parseChapterRange(args[0], bookName)

		for _, chapterID := range chapterIDs {
			fmt.Printf("\n📄 正在审计第%d章...\n", chapterID)
			fmt.Println("────────────────────────────────")

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

			result, err := reviewService.AuditChapter(ctx, bookName, chapterID)
			cancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				continue
			}

			fmt.Printf("✅ 审计完成\n")
			if result != nil {
				fmt.Printf("   章节ID: %d\n", result.ID)
			}
		}
	},
}

// threadCmd represents the thread command
var threadCmd = &cobra.Command{
	Use:   "thread",
	Short: "叙事线程管理",
	Long: `管理多线叙事，包括查看、创建、分配等操作。

子命令:
  list      列出所有线程
  create    创建新线程
  assign    将章节分配到线程
  stats     查看线程统计`,
}

var threadListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出叙事线程",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		threads, err := jsonStore.LoadThreads(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(threads) == 0 {
			fmt.Println("暂无叙事线程")
			return
		}

		fmt.Println("🧵 叙事线程列表")
		fmt.Println("────────────────────────────────")

		for _, thread := range threads {
			fmt.Printf("\n[%s] %s\n", thread.ID, thread.Name)
			fmt.Printf("  类型: %s\n", thread.Type)
			fmt.Printf("  状态: %s\n", thread.Status)
			fmt.Printf("  涉及章节: %v\n", thread.Chapters)
			if thread.LastActiveChapter > 0 {
				fmt.Printf("  最后活跃: 第%d章\n", thread.LastActiveChapter)
			}
		}
	},
}

var threadCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建叙事线程",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		threadType, _ := cmd.Flags().GetString("type")

		if name == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定线程名称 (--name)\n")
			return
		}

		// 加载现有线程
		threads, err := jsonStore.LoadThreads(bookName)
		if err != nil {
			threads = []*model.NarrativeThread{}
		}

		// 创建新线程
		newThread := &model.NarrativeThread{
			ID:               generateID(),
			BookID:           bookName,
			Name:             name,
			Type:             model.ThreadType(threadType),
			Status:           model.ThreadActive,
			Chapters:         []int{},
			LastActiveChapter: 0,
		}

		threads = append(threads, newThread)

		// 保存
		if err := jsonStore.SaveThreads(bookName, threads); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已创建叙事线程: %s\n", name)
		fmt.Printf("   ID: %s\n", truncate(newThread.ID, 8))
	},
}

var threadAssignCmd = &cobra.Command{
	Use:   "assign <线程ID> <章节号>",
	Short: "将章节分配到线程",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		threadID := args[0]
		chapterID := parseChapterID(args[1])

		// 加载线程
		threads, err := jsonStore.LoadThreads(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并更新线程
		found := false
		for _, thread := range threads {
			if thread.ID == threadID || truncate(thread.ID, 8) == threadID {
				thread.Chapters = append(thread.Chapters, chapterID)
				thread.LastActiveChapter = chapterID
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 线程 %s 不存在\n", threadID)
			return
		}

		// 保存
		if err := jsonStore.SaveThreads(bookName, threads); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已将第%d章分配到线程 %s\n", chapterID, truncate(threadID, 8))
	},
}

var threadStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "查看线程统计",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		threads, err := jsonStore.LoadThreads(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		active := 0
		paused := 0
		completed := 0

		for _, thread := range threads {
			switch thread.Status {
			case model.ThreadActive:
				active++
			case model.ThreadPaused:
				paused++
			case model.ThreadComplete:
				completed++
			}
		}

		fmt.Println("🧵 叙事线程统计")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  总数: %d\n", len(threads))
		fmt.Printf("  活跃: %d\n", active)
		fmt.Printf("  暂停: %d\n", paused)
		fmt.Printf("  完成: %d\n", completed)
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(threadCmd)

	threadCmd.AddCommand(threadListCmd)
	threadCmd.AddCommand(threadCreateCmd)
	threadCmd.AddCommand(threadAssignCmd)
	threadCmd.AddCommand(threadStatsCmd)

	threadCreateCmd.Flags().String("name", "", "线程名称")
	threadCreateCmd.Flags().String("type", "main", "线程类型 (main/sub/plot)")
}