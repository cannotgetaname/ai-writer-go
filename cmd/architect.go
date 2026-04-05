package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
	"ai-writer/internal/service"
)

// architectCmd represents the architect command
var architectCmd = &cobra.Command{
	Use:   "architect",
	Short: "架构师工具",
	Long: `AI 架构师工具，用于大纲生成和分形裂变。

子命令:
  generate    生成全书大纲
  fission     分形裂变（展开/优化/分支）
  strategies  查看裂变策略
  analyze     分析结构进度`,
}

var architectGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成全书大纲",
	Long: `根据题材、主角、主题生成全书大纲框架。

示例:
  ai-writer architect generate --genre 玄幻 --main-char "少年天才" --theme 逆袭
  ai-writer architect generate --genre 仙侠 --volumes 5 --target 1000000`,
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}
		genre, _ := cmd.Flags().GetString("genre")
		mainChar, _ := cmd.Flags().GetString("main-char")
		theme, _ := cmd.Flags().GetString("theme")
		targetWords, _ := cmd.Flags().GetInt("target")
		volumes, _ := cmd.Flags().GetInt("volumes")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		architect := service.NewArchitectService(llmClient, jsonStore)
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		result, err := architect.GenerateOutline(ctx, &service.GenerateOutlineRequest{
			BookName:    bookName,
			Genre:       genre,
			MainChar:    mainChar,
			Theme:       theme,
			TargetWords: targetWords,
			VolumeCount: volumes,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Println("【故事梗概】")
		fmt.Println("────────────────────────────────")
		fmt.Println(result.Synopsis)
		fmt.Println()

		fmt.Println("【分卷大纲】")
		fmt.Println("────────────────────────────────")
		for i, vol := range result.Volumes {
			fmt.Printf("\n第%d卷: %s\n", i+1, vol.Label)
			fmt.Printf("  %s\n", vol.Outline)
			if len(vol.Children) > 0 {
				fmt.Println("  章节:")
				for _, ch := range vol.Children {
					fmt.Printf("    - %s: %s\n", ch.Label, ch.Outline)
				}
			}
		}
	},
}

var architectFissionCmd = &cobra.Command{
	Use:   "fission",
	Short: "分形裂变",
	Long: `对大纲节点进行分形裂变操作。

策略:
  expand   展开 - 将简单大纲展开为详细内容
  refine   优化 - 优化现有大纲
  branch   分支 - 生成多条可能的剧情线

示例:
  ai-writer architect fission --strategy expand --count 5 --outline "主角修炼突破"
  ai-writer architect fission --strategy branch --count 3 --outline "最终决战"`,
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}
		strategy, _ := cmd.Flags().GetString("strategy")
		count, _ := cmd.Flags().GetInt("count")
		outline, _ := cmd.Flags().GetString("outline")
		nodeType, _ := cmd.Flags().GetString("type")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		architect := service.NewArchitectService(llmClient, jsonStore)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		result, err := architect.Fission(ctx, &service.FissionRequest{
			BookName:       bookName,
			Strategy:       strategy,
			Count:          count,
			CurrentOutline: outline,
			NodeType:       nodeType,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Printf("【分形裂变结果 - %s】\n", strategy)
		fmt.Println("────────────────────────────────")
		for i, node := range result.Nodes {
			fmt.Printf("\n%d. %s\n", i+1, node.Label)
			fmt.Printf("   状态: %s\n", node.Status)
			fmt.Printf("   概述: %s\n", node.Outline)
		}
	},
}

var architectStrategiesCmd = &cobra.Command{
	Use:   "strategies",
	Short: "查看裂变策略",
	Run: func(cmd *cobra.Command, args []string) {
		strategies := service.GetFissionStrategies()

		fmt.Println("【分形裂变策略】")
		fmt.Println("═══════════════════════════════════════")

		for category, items := range strategies {
			fmt.Printf("\n%s 策略:\n", category)
			for _, s := range items {
				fmt.Printf("  %-15s - %s\n", s.Name, s.Description)
			}
		}
	},
}

var architectAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "分析结构进度",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

		// 加载书籍结构
		chapters, err := jsonStore.LoadChapters(bookName)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		volumes, err := jsonStore.LoadVolumes(bookName)
		if err != nil {
			volumes = []*model.Volume{}
		}

		// 统计
		total := len(chapters)
		done := 0
		totalWords := 0

		for _, ch := range chapters {
			content, err := jsonStore.LoadChapterContent(bookName, ch.ID)
			if err == nil && len(content) > 1000 {
				done++
			}
			totalWords += ch.WordCount
		}

		progress := 0.0
		if total > 0 {
			progress = float64(done) / float64(total) * 100
		}

		fmt.Println("【结构分析】")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("分卷数量: %d\n", len(volumes))
		fmt.Printf("章节总数: %d\n", total)
		fmt.Printf("已完成:   %d (%.1f%%)\n", done, progress)
		fmt.Printf("总字数:   %d\n", totalWords)
	},
}

func init() {
	rootCmd.AddCommand(architectCmd)

	architectCmd.AddCommand(architectGenerateCmd)
	architectCmd.AddCommand(architectFissionCmd)
	architectCmd.AddCommand(architectStrategiesCmd)
	architectCmd.AddCommand(architectAnalyzeCmd)

	// generate 命令选项
	architectGenerateCmd.Flags().String("genre", "玄幻", "题材类型")
	architectGenerateCmd.Flags().String("main-char", "", "主角设定")
	architectGenerateCmd.Flags().String("theme", "", "故事主题")
	architectGenerateCmd.Flags().Int("target", 1000000, "目标字数")
	architectGenerateCmd.Flags().Int("volumes", 3, "分卷数量")

	// fission 命令选项
	architectFissionCmd.Flags().String("strategy", "expand", "裂变策略 (expand/refine/branch)")
	architectFissionCmd.Flags().Int("count", 5, "生成数量")
	architectFissionCmd.Flags().String("outline", "", "当前大纲")
	architectFissionCmd.Flags().String("type", "volume", "节点类型 (volume/chapter)")
}