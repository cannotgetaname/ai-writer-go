package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/service"
)

// analysisCmd represents the analysis command
var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "拆书分析",
	Long: `拆书分析工具，用于分析优秀作品学习写作技巧。

子命令:
  parse     解析TXT文件
  analyze   分析作品内容
  outline   提取作品大纲
  compare   与自己的作品对比`,
}

var analysisParseCmd = &cobra.Command{
	Use:   "parse <文件路径>",
	Short: "解析TXT文件",
	Long: `解析TXT文件，自动识别章节结构。

示例:
  ai-writer analysis parse novel.txt
  ai-writer analysis parse /path/to/book.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// 读取文件
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "读取文件失败: %v\n", err)
			return
		}

		// 解析
		analysis := service.NewAnalysisService(nil, nil)
		result := analysis.ParseTXT(string(content))

		if !result.Success {
			fmt.Fprintf(cmd.OutOrStderr(), "解析失败: %s\n", result.Message)
			return
		}

		fmt.Println("【解析结果】")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("章节数量: %d\n", result.ChapterCount)
		fmt.Printf("总字数:   %d\n", result.TotalWords)
		fmt.Println()

		fmt.Println("【章节列表】")
		fmt.Println("────────────────────────────────")
		for _, ch := range result.Chapters {
			if len(ch.Title) > 30 {
				ch.Title = ch.Title[:30] + "..."
			}
			fmt.Printf("第%d章: %s (%d字)\n", ch.Num, ch.Title, ch.WordCount)
		}
	},
}

var analysisAnalyzeCmd = &cobra.Command{
	Use:   "analyze <文件路径>",
	Short: "分析作品内容",
	Long: `使用AI分析作品内容，提取人物、剧情、世界观等。

示例:
  ai-writer analysis analyze novel.txt
  ai-writer analysis analyze book.txt --type full`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		analysisType, _ := cmd.Flags().GetString("type")

		// 读取文件
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "读取文件失败: %v\n", err)
			return
		}

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		// 解析
		analysisSvc := service.NewAnalysisService(llmClient, jsonStore)
		parseResult := analysisSvc.ParseTXT(string(content))

		if !parseResult.Success {
			fmt.Fprintf(cmd.OutOrStderr(), "解析失败: %s\n", parseResult.Message)
			return
		}

		fmt.Printf("正在分析 %d 章内容...\n", parseResult.ChapterCount)

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		result, err := analysisSvc.AnalyzeChapters(ctx, parseResult.Chapters, analysisType)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "分析失败: %v\n", err)
			return
		}

		fmt.Println("\n【分析结果】")
		fmt.Println("═══════════════════════════════════════")

		fmt.Println("\n【内容摘要】")
		fmt.Println(result.Summary)

		fmt.Println("\n【世界观设定】")
		fmt.Println(result.WorldSetting)

		fmt.Println("\n【写作风格】")
		fmt.Println(result.WritingStyle)

		if len(result.Characters) > 0 {
			fmt.Println("\n【主要人物】")
			for _, c := range result.Characters {
				fmt.Printf("  - %s (%s): %s\n", c.Name, c.Role, c.Description)
			}
		}

		if len(result.PlotPoints) > 0 {
			fmt.Println("\n【剧情点】")
			for _, p := range result.PlotPoints {
				fmt.Printf("  第%d章 [%s]: %s\n", p.Chapter, p.Type, p.Description)
			}
		}
	},
}

var analysisOutlineCmd = &cobra.Command{
	Use:   "outline <文件路径>",
	Short: "提取作品大纲",
	Long: `从作品中提取大纲结构。

示例:
  ai-writer analysis outline novel.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// 读取文件
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "读取文件失败: %v\n", err)
			return
		}

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		// 解析
		analysisSvc := service.NewAnalysisService(llmClient, jsonStore)
		parseResult := analysisSvc.ParseTXT(string(content))

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		result, err := analysisSvc.ExtractOutline(ctx, parseResult.Chapters)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "提取失败: %v\n", err)
			return
		}

		fmt.Println("【作品大纲】")
		fmt.Println("═══════════════════════════════════════")

		for i, vol := range result {
			fmt.Printf("\n第%d卷: %s\n", i+1, vol.Label)
			fmt.Printf("  %s\n", vol.Outline)
			if len(vol.Children) > 0 {
				for _, ch := range vol.Children {
					fmt.Printf("  - %s: %s\n", ch.Label, ch.Outline)
				}
			}
		}
	},
}

var analysisCompareCmd = &cobra.Command{
	Use:   "compare <文件路径>",
	Short: "与自己的作品对比",
	Long: `将参考作品与当前作品进行对比分析。

示例:
  ai-writer analysis compare reference.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}
		filePath := args[0]

		// 读取文件
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "读取文件失败: %v\n", err)
			return
		}

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		// 解析和分析
		analysisSvc := service.NewAnalysisService(llmClient, jsonStore)
		parseResult := analysisSvc.ParseTXT(string(content))

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		analysis, err := analysisSvc.AnalyzeChapters(ctx, parseResult.Chapters, "full")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "分析失败: %v\n", err)
			return
		}

		fmt.Printf("正在与《%s》对比分析...\n", bookName)

		result, err := analysisSvc.CompareWithMyWork(ctx, analysis, bookName)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "对比失败: %v\n", err)
			return
		}

		fmt.Println("\n【对比分析】")
		fmt.Println("═══════════════════════════════════════")
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(analysisCmd)

	analysisCmd.AddCommand(analysisParseCmd)
	analysisCmd.AddCommand(analysisAnalyzeCmd)
	analysisCmd.AddCommand(analysisOutlineCmd)
	analysisCmd.AddCommand(analysisCompareCmd)

	analysisAnalyzeCmd.Flags().String("type", "full", "分析类型 (full/character/plot/style)")
}