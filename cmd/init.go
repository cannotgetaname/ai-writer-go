package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/service"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "从创意初始化项目",
	Long: `从创意点子一键初始化整个小说项目，包括世界观、主角、大纲。

AI 会自动生成：
1. 世界观设定（力量体系、势力分布、特殊规则等）
2. 主角设定（性格、目标、背景等）
3. 完整大纲（分卷 → 章节标题 + 章节大纲）

示例:
  ai-writer init --idea "少年获得系统在修仙世界逆袭" --genre 玄幻 --name "我的小说"
  ai-writer init --idea "都市重生商战" --genre 都市 --name "商业帝国" --volumes 3`,
	Run: func(cmd *cobra.Command, args []string) {
		idea, _ := cmd.Flags().GetString("idea")
		genre, _ := cmd.Flags().GetString("genre")
		name, _ := cmd.Flags().GetString("name")
		targetWords, _ := cmd.Flags().GetInt("target")
		volumeCount, _ := cmd.Flags().GetInt("volumes")

		if idea == "" {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: 请指定创意 (--idea)\n")
			return
		}

		if name == "" {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: 请指定书名 (--name)\n")
			return
		}

		// 检查书籍是否已存在
		books, _ := jsonStore.ListBooks()
		for _, b := range books {
			if b.Name == name {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: 书籍 '%s' 已存在\n", name)
				return
			}
		}

		// 初始化 LLM 客户端
		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		// 创建初始化服务
		initService := service.NewInitService(llmClient, jsonStore)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		fmt.Println("🚀 开始初始化项目...")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  书名: %s\n", name)
		fmt.Printf("  题材: %s\n", genre)
		fmt.Printf("  创意: %s\n", truncate(idea, 50))
		fmt.Println("────────────────────────────────")

		// 创建书籍目录
		_, err = jsonStore.CreateBook(name)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: 创建书籍失败: %v\n", err)
			return
		}

		// 执行初始化
		result, err := initService.Initialize(ctx, &service.InitRequest{
			BookName:    name,
			Idea:        idea,
			Genre:       genre,
			TargetWords: targetWords,
			VolumeCount: volumeCount,
		})

		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			// 清理已创建的书籍
			jsonStore.DeleteBook(name)
			return
		}

		// 保存结果
		if err := initService.SaveInitResult(result); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: 保存失败: %v\n", err)
			return
		}

		// 输出结果
		fmt.Println("\n✅ 项目初始化完成！")
		fmt.Println("═══════════════════════════════════════")

		fmt.Println("\n📖 故事梗概")
		fmt.Println("────────────────────────────────")
		fmt.Println(result.Synopsis)

		fmt.Println("\n🌍 世界观")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  题材: %s\n", result.WorldView.BasicInfo.Genre)
		fmt.Printf("  力量体系: %s\n", truncate(result.WorldView.CoreSettings.PowerSystem, 50))
		fmt.Printf("  社会结构: %s\n", truncate(result.WorldView.CoreSettings.SocialStructure, 50))

		if result.MainCharacter != nil {
			fmt.Println("\n👤 主角")
			fmt.Println("────────────────────────────────")
			fmt.Printf("  姓名: %s\n", result.MainCharacter.Name)
			fmt.Printf("  性别: %s\n", result.MainCharacter.Gender)
			fmt.Printf("  简介: %s\n", result.MainCharacter.Bio)
		}

		fmt.Println("\n📚 大纲")
		fmt.Println("────────────────────────────────")

		totalChapters := 0
		for i, vol := range result.Volumes {
			chapterCount := len(vol.Chapters)
			totalChapters += chapterCount
			fmt.Printf("\n  第%d卷: %s (%d章)\n", i+1, vol.Title, chapterCount)
			if verbose && len(vol.Chapters) > 0 {
				for _, ch := range vol.Chapters {
					fmt.Printf("    第%d章: %s\n", ch.ID, ch.Title)
				}
			}
		}

		fmt.Printf("\n────────────────────────────────")
		fmt.Printf("\n  共 %d 卷, %d 章\n", len(result.Volumes), totalChapters)

		fmt.Println("\n使用以下命令继续创作:")
		fmt.Println("────────────────────────────────")
		fmt.Printf("  查看大纲: ai-writer -b %s chapter list\n", name)
		fmt.Printf("  生成章节: ai-writer -b %s write 1 --stream\n", name)
		fmt.Printf("  批量生成: ai-writer -b %s batch generate --from 1 --to %d\n", name, totalChapters)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("idea", "", "故事创意/核心点子")
	initCmd.Flags().String("genre", "玄幻", "题材类型 (玄幻/仙侠/都市/科幻等)")
	initCmd.Flags().String("name", "", "书名")
	initCmd.Flags().Int("target", 1000000, "目标字数")
	initCmd.Flags().Int("volumes", 5, "分卷数量")
}