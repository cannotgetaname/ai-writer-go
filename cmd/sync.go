package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/service"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "状态同步",
	Long: `从章节提取状态变更，审核后应用到项目。

流程:
1. extract - 从章节提取状态变更
2. review  - 查看待审核变更
3. apply   - 应用审核后的变更

示例:
  ai-writer sync extract 1          # 从第1章提取状态变更
  ai-writer sync pending            # 查看待审核变更
  ai-writer sync apply              # 应用所有变更
  ai-writer sync apply --change id  # 应用指定变更`,
}

var syncExtractCmd = &cobra.Command{
	Use:   "extract <章节号>",
	Short: "从章节提取状态变更",
	Args:  cobra.ExactArgs(1),
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

		syncService := service.NewSyncService(llmClient, jsonStore)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		fmt.Printf("正在从第%d章提取状态变更...\n", chapterID)

		pending, err := syncService.ExtractStateChanges(ctx, bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(pending.Changes) == 0 {
			fmt.Println("✅ 未检测到状态变更")
			return
		}

		// 保存待审核变更
		if err := savePendingChanges(bookName, pending); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 提取完成，共 %d 条变更待审核\n", len(pending.Changes))
		fmt.Println("\n使用 'ai-writer sync pending' 查看详情")
	},
}

var syncPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "查看待审核变更",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		pending := loadPendingChanges(bookName)
		if pending == nil || len(pending.Changes) == 0 {
			fmt.Println("暂无待审核变更")
			return
		}

		fmt.Println("📋 待审核状态变更")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("来源: 第%d章\n", pending.ChapterID)
		fmt.Printf("提取时间: %s\n", pending.ExtractedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("────────────────────────────────")

		for i, change := range pending.Changes {
			fmt.Printf("\n[%d] %s\n", i+1, change.Type)
			fmt.Printf("    实体: %s\n", change.Entity)
			if change.Field != "" {
				fmt.Printf("    字段: %s\n", change.Field)
			}
			if change.OldValue != "" {
				fmt.Printf("    旧值: %s\n", change.OldValue)
			}
			fmt.Printf("    新值: %s\n", change.NewValue)
			if change.Reason != "" {
				fmt.Printf("    原因: %s\n", change.Reason)
			}
			fmt.Printf("    ID: %s\n", truncate(change.ID, 8))
		}

		fmt.Println("\n────────────────────────────────")
		fmt.Println("使用 'ai-writer sync apply' 应用变更")
		fmt.Println("使用 'ai-writer sync reject' 丢弃变更")
	},
}

var syncApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "应用状态变更",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		pending := loadPendingChanges(bookName)
		if pending == nil || len(pending.Changes) == 0 {
			fmt.Println("暂无待审核变更")
			return
		}

		changeID, _ := cmd.Flags().GetString("change")

		// 初始化同步服务
		syncService := service.NewSyncService(nil, jsonStore)

		applied := 0
		for _, change := range pending.Changes {
			// 如果指定了变更ID，只应用该变更
			if changeID != "" && truncate(change.ID, 8) != changeID && change.ID != changeID {
				continue
			}

			if err := syncService.ApplyChange(bookName, &change); err != nil {
				fmt.Printf("❌ [%s] %s: %v\n", truncate(change.ID, 8), change.Entity, err)
				continue
			}

			fmt.Printf("✅ [%s] %s: %s → %s\n", truncate(change.ID, 8), change.Entity, change.OldValue, change.NewValue)
			applied++
		}

		// 清除待审核变更
		if changeID == "" {
			deletePendingChanges(bookName)
		}

		fmt.Printf("\n已应用 %d 条变更\n", applied)
	},
}

var syncRejectCmd = &cobra.Command{
	Use:   "reject",
	Short: "丢弃状态变更",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Print("确定要丢弃所有待审核变更吗？输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		deletePendingChanges(bookName)
		fmt.Println("✅ 已丢弃所有待审核变更")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.AddCommand(syncExtractCmd)
	syncCmd.AddCommand(syncPendingCmd)
	syncCmd.AddCommand(syncApplyCmd)
	syncCmd.AddCommand(syncRejectCmd)

	syncApplyCmd.Flags().String("change", "", "指定要应用的变更ID")
	syncRejectCmd.Flags().BoolP("force", "f", false, "强制丢弃，不确认")
}

// savePendingChanges 保存待审核变更
func savePendingChanges(bookName string, pending *service.PendingChanges) error {
	return printJSONToFile(pending, getPendingChangesFile(bookName))
}

// loadPendingChanges 加载待审核变更
func loadPendingChanges(bookName string) *service.PendingChanges {
	data, err := os.ReadFile(getPendingChangesFile(bookName))
	if err != nil {
		return nil
	}

	var pending service.PendingChanges
	if err := parseJSONFromBytes(data, &pending); err != nil {
		return nil
	}

	return &pending
}

// deletePendingChanges 删除待审核变更
func deletePendingChanges(bookName string) {
	os.Remove(getPendingChangesFile(bookName))
}

// getPendingChangesFile 获取待审核变更文件路径
func getPendingChangesFile(bookName string) string {
	return fmt.Sprintf("data/projects/%s/.pending_changes.json", bookName)
}