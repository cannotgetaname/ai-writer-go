package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// foreshadowCmd represents the foreshadow command
var foreshadowCmd = &cobra.Command{
	Use:   "foreshadow",
	Short: "伏笔管理",
	Long: `管理书籍伏笔，包括查看、添加、回收等操作。

子命令:
  list      列出所有伏笔
  add       添加伏笔
  resolve   回收伏笔
  warnings  查看伏笔预警`,
}

var foreshadowListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出伏笔",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		statusFilter, _ := cmd.Flags().GetString("status")
		typeFilter, _ := cmd.Flags().GetString("type")

		foreshadows, err := jsonStore.LoadForeshadows(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(foreshadows) == 0 {
			fmt.Println("暂无伏笔")
			return
		}

		// Filter
		var filtered []*model.Foreshadow
		for _, f := range foreshadows {
			if statusFilter != "" && string(f.Status) != statusFilter {
				continue
			}
			if typeFilter != "" && string(f.Type) != typeFilter {
				continue
			}
			filtered = append(filtered, f)
		}

		switch outputFormat {
		case "json":
			printJSON(filtered)
		case "markdown":
			fmt.Println("# 伏笔列表")
			fmt.Println()
			for _, f := range filtered {
				fmt.Printf("- **%s** [%s] - 第%d章埋设\n", f.Content, f.Status, f.SourceChapter)
			}
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "内容\t类型\t状态\t埋设章节\t回收章节")
			fmt.Fprintln(w, "----\t----\t----\t--------\t--------")
			for _, f := range filtered {
				resolved := "-"
				if f.ResolvedChapter > 0 {
					resolved = fmt.Sprintf("%d", f.ResolvedChapter)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
					truncate(f.Content, 30), f.Type, f.Status, f.SourceChapter, resolved)
			}
			w.Flush()
		}
	},
}

var foreshadowAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加伏笔",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		content, _ := cmd.Flags().GetString("content")
		fsType, _ := cmd.Flags().GetString("type")
		chapter, _ := cmd.Flags().GetInt("chapter")
		target, _ := cmd.Flags().GetInt("target")
		importance, _ := cmd.Flags().GetString("importance")

		if content == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定伏笔内容 (--content)\n")
			return
		}

		// 加载现有伏笔
		foreshadows, err := jsonStore.LoadForeshadows(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 创建新伏笔
		now := time.Now()
		newForeshadow := &model.Foreshadow{
			ID:            generateID(),
			BookID:        bookName,
			Content:       content,
			Type:          model.ForeshadowType(fsType),
			Importance:    model.Importance(importance),
			SourceChapter: chapter,
			TargetChapter: target,
			Status:        model.ForeshadowActive,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		foreshadows = append(foreshadows, newForeshadow)

		// 保存
		if err := jsonStore.SaveForeshadows(bookName, foreshadows); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已添加伏笔: %s\n", truncate(newForeshadow.ID, 8))
		if verbose {
			fmt.Printf("   内容: %s\n", content)
			fmt.Printf("   类型: %s\n", fsType)
			fmt.Printf("   埋设章节: %d\n", chapter)
			if target > 0 {
				fmt.Printf("   预期回收章节: %d\n", target)
			}
		}
	},
}

var foreshadowResolveCmd = &cobra.Command{
	Use:   "resolve <伏笔ID>",
	Short: "回收伏笔",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		fsID := args[0]
		chapter, _ := cmd.Flags().GetInt("chapter")
		resolvedContent, _ := cmd.Flags().GetString("resolved-content")

		// 加载伏笔列表
		foreshadows, err := jsonStore.LoadForeshadows(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并更新伏笔
		found := false
		for _, fs := range foreshadows {
			if fs.ID == fsID || truncate(fs.ID, 8) == fsID {
				fs.Status = model.ForeshadowResolved
				fs.ResolvedChapter = chapter
				fs.ResolvedContent = resolvedContent
				fs.UpdatedAt = time.Now()
				fsID = fs.ID
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 伏笔 %s 不存在\n", fsID)
			return
		}

		// 保存
		if err := jsonStore.SaveForeshadows(bookName, foreshadows); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已回收伏笔: %s\n", truncate(fsID, 8))
		fmt.Printf("   回收章节: %d\n", chapter)
		if resolvedContent != "" {
			fmt.Printf("   回收内容: %s\n", resolvedContent)
		}
	},
}

var foreshadowWarningsCmd = &cobra.Command{
	Use:   "warnings",
	Short: "查看伏笔预警",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 加载伏笔
		foreshadows, err := jsonStore.LoadForeshadows(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 加载章节获取当前最新章节
		chapters, err := jsonStore.LoadChapters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		currentChapter := 0
		for _, ch := range chapters {
			if ch.ID > currentChapter {
				currentChapter = ch.ID
			}
		}

		// 检查预警
		const warningThreshold = 5 // 超过5章未回收则预警
		var warnings []string

		for _, fs := range foreshadows {
			if fs.Status != model.ForeshadowActive {
				continue
			}

			gap := currentChapter - fs.SourceChapter
			if gap > warningThreshold {
				warnings = append(warnings, fmt.Sprintf(
					"⚠️  [%s] %s (埋设: 第%d章, 已过 %d 章)",
					truncate(fs.ID, 8), truncate(fs.Content, 30), fs.SourceChapter, gap,
				))
			}

			// 检查是否超过预期回收章节
			if fs.TargetChapter > 0 && currentChapter > fs.TargetChapter {
				warnings = append(warnings, fmt.Sprintf(
					"⏰ [%s] %s (预期: 第%d章, 当前: 第%d章)",
					truncate(fs.ID, 8), truncate(fs.Content, 30), fs.TargetChapter, currentChapter,
				))
			}
		}

		fmt.Println("伏笔预警:")
		fmt.Println("────────────────────────────────")

		if len(warnings) == 0 {
			fmt.Println("  ✅ 暂无预警信息")
		} else {
			for _, w := range warnings {
				fmt.Printf("  %s\n", w)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(foreshadowCmd)

	foreshadowCmd.AddCommand(foreshadowListCmd)
	foreshadowCmd.AddCommand(foreshadowAddCmd)
	foreshadowCmd.AddCommand(foreshadowResolveCmd)
	foreshadowCmd.AddCommand(foreshadowWarningsCmd)

	foreshadowListCmd.Flags().StringP("status", "s", "", "筛选状态 (active/resolved)")
	foreshadowListCmd.Flags().StringP("type", "t", "", "筛选类型 (剧情/人物/物品)")

	foreshadowAddCmd.Flags().String("content", "", "伏笔内容")
	foreshadowAddCmd.Flags().String("type", "plot", "伏笔类型 (plot/character/item/mystery)")
	foreshadowAddCmd.Flags().Int("chapter", 1, "埋设章节")
	foreshadowAddCmd.Flags().Int("target", 0, "预期回收章节")
	foreshadowAddCmd.Flags().String("importance", "medium", "重要程度 (high/medium/low)")

	foreshadowResolveCmd.Flags().Int("chapter", 0, "回收章节")
	foreshadowResolveCmd.Flags().String("resolved-content", "", "回收内容")
}