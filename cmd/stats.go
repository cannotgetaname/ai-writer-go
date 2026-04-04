package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/engine"
	"ai-writer/internal/model"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "统计信息",
	Long: `查看书籍统计信息，包括字数、章节、人物等。`,
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 加载书籍信息
		volumes, _ := jsonStore.LoadVolumes(bookName)
		chapters, _ := jsonStore.LoadChapters(bookName)
		characters, _ := jsonStore.LoadCharacters(bookName)
		items, _ := jsonStore.LoadItems(bookName)
		locations, _ := jsonStore.LoadLocations(bookName)
		foreshadows, _ := jsonStore.LoadForeshadows(bookName)

		fmt.Printf("📊 %s 统计\n", bookName)
		fmt.Println("────────────────────────────────")

		fmt.Printf("  分卷数:   %d\n", len(volumes))
		fmt.Printf("  章节数:   %d\n", len(chapters))

		// 统计字数
		var totalWords, completedChapters int
		for _, ch := range chapters {
			content, err := jsonStore.LoadChapterContent(bookName, ch.ID)
			if err == nil {
				wordCount := countChineseChars(content)
				totalWords += wordCount
				if wordCount > 0 {
					completedChapters++
				}
			}
		}

		fmt.Printf("  总字数:   %d\n", totalWords)
		fmt.Printf("  完成章节: %d / %d\n", completedChapters, len(chapters))
		fmt.Println()
		fmt.Printf("  人物数:   %d\n", len(characters))
		fmt.Printf("  物品数:   %d\n", len(items))
		fmt.Printf("  地点数:   %d\n", len(locations))
		fmt.Println()

		// 伏笔统计
		var active, resolved int
		for _, f := range foreshadows {
			if f.Status == "active" {
				active++
			} else if f.Status == "resolved" {
				resolved++
			}
		}
		fmt.Printf("  伏笔:     %d (活跃: %d, 已回收: %d)\n", len(foreshadows), active, resolved)

		// 详细模式
		if verbose {
			fmt.Println()
			fmt.Println("章节详情:")
			fmt.Println("────────────────────────────────")
			for _, ch := range chapters {
				content, _ := jsonStore.LoadChapterContent(bookName, ch.ID)
				wordCount := countChineseChars(content)
				status := "❌ 未写"
				if wordCount > 0 {
					status = "✅ 已写"
				}
				fmt.Printf("  第%d章 %s: %d字 %s\n", ch.ID, ch.Title, wordCount, status)
			}
		}
	},
}

// timelineCmd represents the timeline command
var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "时间线",
	Long: `查看故事时间线。`,
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapters, err := jsonStore.LoadChapters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		fmt.Println("📅 时间线")
		fmt.Println("────────────────────────────────")

		for _, ch := range chapters {
			fmt.Printf("第%d章: %s\n", ch.ID, ch.Title)
			if ch.TimeInfo.Label != "" {
				fmt.Printf("  时间: %s\n", ch.TimeInfo.Label)
			}
			if ch.TimeInfo.Duration != "" && ch.TimeInfo.Duration != "0" {
				fmt.Printf("  时长: %s\n", ch.TimeInfo.Duration)
			}
			if len(ch.TimeInfo.Events) > 0 {
				fmt.Println("  事件:")
				for _, event := range ch.TimeInfo.Events {
					fmt.Printf("    - %s\n", event)
				}
			}
			fmt.Println()
		}
	},
}

// causalCmd represents the causal chain command
var causalCmd = &cobra.Command{
	Use:   "causal",
	Short: "因果链",
	Long: `查看因果链追踪信息。`,
}

var causalShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示因果链",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		events, err := jsonStore.LoadCausalChains(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(events) == 0 {
			fmt.Println("暂无因果链数据")
			return
		}

		fmt.Println("🔗 因果链")
		fmt.Println("────────────────────────────────")

		for _, event := range events {
			fmt.Printf("第%d章\n", event.ChapterID)
			fmt.Printf("  因: %s\n", event.Cause)
			fmt.Printf("  事: %s\n", event.Event)
			fmt.Printf("  果: %s\n", event.Effect)
			fmt.Printf("  决: %s\n", event.Decision)
			fmt.Println()
		}
	},
}

func init() {
		rootCmd.AddCommand(statsCmd)
		rootCmd.AddCommand(timelineCmd)
		rootCmd.AddCommand(causalCmd)

		causalCmd.AddCommand(causalShowCmd)
		causalCmd.AddCommand(causalExtractCmd)
		causalCmd.AddCommand(causalValidateCmd)
	}

	// causalExtractCmd 从章节提取因果链
	var causalExtractCmd = &cobra.Command{
		Use:   "extract <章节号>",
		Short: "从章节提取因果链",
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

			// 创建因果链引擎
			causalEngine := engine.NewCausalChainEngine(llmClient, jsonStore)

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			fmt.Printf("正在从第%d章提取因果链...\n", chapterID)

			event, err := causalEngine.ExtractFromChapter(ctx, bookName, chapterID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				return
			}

			// 加载现有因果链
			events, err := jsonStore.LoadCausalChains(bookName)
			if err != nil {
				events = []*model.CausalEvent{}
			}

			// 检查是否已存在
			for i, e := range events {
				if e.ChapterID == chapterID {
					events[i] = event
					break
				}
			}

			// 如果不存在则添加
			found := false
			for _, e := range events {
				if e.ChapterID == chapterID {
					found = true
					break
				}
			}
			if !found {
				events = append(events, event)
			}

			// 保存
			if err := jsonStore.SaveCausalChains(bookName, events); err != nil {
				fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
				return
			}

			fmt.Println("✅ 因果链提取完成")
			fmt.Println("────────────────────────────────")
			fmt.Printf("  因: %s\n", event.Cause)
			fmt.Printf("  事: %s\n", event.Event)
			fmt.Printf("  果: %s\n", event.Effect)
			fmt.Printf("  决: %s\n", event.Decision)
		},
	}

	// causalValidateCmd 验证因果链一致性
	var causalValidateCmd = &cobra.Command{
		Use:   "validate",
		Short: "验证因果链一致性",
		Run: func(cmd *cobra.Command, args []string) {
			bookName, err := requireBookName()
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				return
			}

			// 创建因果链引擎
			causalEngine := engine.NewCausalChainEngine(nil, jsonStore)

			issues := causalEngine.ValidateChain(bookName)

			fmt.Println("🔍 因果链验证结果")
			fmt.Println("────────────────────────────────")

			if len(issues) == 0 {
				fmt.Println("  ✅ 因果链完整，无问题")
			} else {
				fmt.Println("  ⚠️  发现以下问题:")
				for _, issue := range issues {
					fmt.Printf("    - %s\n", issue)
				}
			}
		},
	}