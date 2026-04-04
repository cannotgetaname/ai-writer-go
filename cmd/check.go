package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/engine"
)

// checkCmd represents the consistency check command
var checkCmd = &cobra.Command{
	Use:   "check <章节号>",
	Short: "一致性检查",
	Long: `检查章节内容与项目设定的一致性。

检查维度:
1. 人设一致性 - 人物言行是否符合设定
2. 物品归属 - 物品持有者是否正确
3. 地点描述 - 地点信息是否一致
4. 剧情逻辑 - 是否有前后矛盾
5. 时间线 - 时间推进是否合理

示例:
  ai-writer check 1              # 检查第1章
  ai-writer check 1-10           # 检查第1到10章
  ai-writer check all            # 检查所有章节`,
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

		// 创建一致性检查引擎
		consistencyEngine := engine.NewConsistencyEngine(llmClient, jsonStore)

		// 解析章节范围
		chapterIDs := parseChapterRange(args[0], bookName)

		fmt.Printf("\n🔍 一致性检查: %s\n", bookName)
		fmt.Println("═══════════════════════════════════════")

		var allIssues []engine.ConsistencyIssue

		for _, chapterID := range chapterIDs {
			fmt.Printf("\n检查第%d章...\n", chapterID)

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

			report, err := consistencyEngine.CheckChapter(ctx, bookName, chapterID)
			cancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "  ❌ 检查失败: %v\n", err)
				continue
			}

			if len(report.Issues) == 0 {
				fmt.Println("  ✅ 无问题")
			} else {
				fmt.Printf("  ⚠️  发现 %d 个问题\n", len(report.Issues))
				allIssues = append(allIssues, report.Issues...)
			}
		}

		// 显示汇总报告
		fmt.Println("\n═══════════════════════════════════════")
		fmt.Println("📊 检查汇总")
		fmt.Println("────────────────────────────────")

		if len(allIssues) == 0 {
			fmt.Println("✅ 所有检查章节未发现一致性问题")
			return
		}

		// 按严重程度统计
		highCount := 0
		mediumCount := 0
		lowCount := 0

		for _, issue := range allIssues {
			switch issue.Severity {
			case "high":
				highCount++
			case "medium":
				mediumCount++
			default:
				lowCount++
			}
		}

		fmt.Printf("总计: %d 个问题\n", len(allIssues))
		fmt.Printf("  严重: %d\n", highCount)
		fmt.Printf("  中等: %d\n", mediumCount)
		fmt.Printf("  轻微: %d\n", lowCount)

		// 显示详细问题
		fmt.Println("\n问题详情:")
		fmt.Println("────────────────────────────────")

		for i, issue := range allIssues {
			fmt.Printf("\n[%d] [%s][%s] %s\n", i+1, issue.Type, issue.Severity, issue.Description)
			if issue.Entity != "" {
				fmt.Printf("    实体: %s\n", issue.Entity)
			}
			if issue.Suggestion != "" {
				fmt.Printf("    建议: %s\n", issue.Suggestion)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}