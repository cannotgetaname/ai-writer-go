package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// paragraphCmd represents the paragraph command
var paragraphCmd = &cobra.Command{
	Use:   "paragraph",
	Short: "段落管理",
	Long: `管理章节段落，包括查看、添加、编辑、删除等操作。

子命令:
  list      列出章节所有段落
  add       添加段落
  edit      编辑段落
  delete    删除段落
  move      移动段落位置`,
}

var paragraphListCmd = &cobra.Command{
	Use:   "list <章节号>",
	Short: "列出段落",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])

		paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(paragraphs.Paragraphs) == 0 {
			fmt.Println("暂无段落")
			return
		}

		switch outputFormat {
		case "json":
			printJSON(paragraphs)
		default:
			fmt.Printf("📖 第%d章段落列表\n", chapterID)
			fmt.Println("────────────────────────────────")
			fmt.Printf("段落数: %d  总字数: %d\n", paragraphs.Metadata.ParagraphCount, paragraphs.Metadata.TotalWords)
			fmt.Println("────────────────────────────────")

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "序号\tID\t字数\t预览")
			fmt.Fprintln(w, "----\t----\t----\t----")

			for i, p := range paragraphs.Paragraphs {
				preview := truncate(p.Text, 30)
				fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", i+1, truncate(p.ID, 8), p.WordCount, preview)
			}
			w.Flush()
		}
	},
}

var paragraphShowCmd = &cobra.Command{
	Use:   "show <章节号> <段落ID>",
	Short: "查看段落内容",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		paragraphID := args[1]

		paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		for _, p := range paragraphs.Paragraphs {
			if p.ID == paragraphID || truncate(p.ID, 8) == paragraphID {
				fmt.Printf("📝 段落 %s\n", p.ID)
				fmt.Println("────────────────────────────────")
				fmt.Printf("字数: %d\n", p.WordCount)
				fmt.Println("────────────────────────────────")
				fmt.Println(p.Text)
				return
			}
		}

		fmt.Fprintf(os.Stderr, "错误: 段落 %s 不存在\n", paragraphID)
	},
}

var paragraphAddCmd = &cobra.Command{
	Use:   "add <章节号>",
	Short: "添加段落",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		text, _ := cmd.Flags().GetString("text")
		position, _ := cmd.Flags().GetInt("position")

		if text == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定段落内容 (--text)\n")
			return
		}

		// 创建新段落
		newParagraph := &model.Paragraph{
			ID:        generateID(),
			Text:      text,
			WordCount: countChineseChars(text),
		}

		// 加载现有段落
		paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 添加到指定位置
		if position < 0 || position >= len(paragraphs.Paragraphs) {
			paragraphs.Paragraphs = append(paragraphs.Paragraphs, *newParagraph)
		} else {
			paragraphs.Paragraphs = append(
				paragraphs.Paragraphs[:position],
				append([]model.Paragraph{*newParagraph}, paragraphs.Paragraphs[position:]...)...,
			)
		}

		// 保存
		if err := jsonStore.SaveChapterParagraphs(bookName, paragraphs); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已添加段落: %s\n", truncate(newParagraph.ID, 8))
		fmt.Printf("   字数: %d\n", newParagraph.WordCount)
	},
}

var paragraphEditCmd = &cobra.Command{
	Use:   "edit <章节号> <段落ID>",
	Short: "编辑段落",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		paragraphID := args[1]
		text, _ := cmd.Flags().GetString("text")

		if text == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定段落内容 (--text)\n")
			return
		}

		// 加载段落
		paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并更新
		found := false
		for i, p := range paragraphs.Paragraphs {
			if p.ID == paragraphID || truncate(p.ID, 8) == paragraphID {
				paragraphs.Paragraphs[i].Text = text
				paragraphs.Paragraphs[i].WordCount = countChineseChars(text)
				paragraphID = p.ID
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 段落 %s 不存在\n", paragraphID)
			return
		}

		// 保存
		if err := jsonStore.SaveChapterParagraphs(bookName, paragraphs); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已更新段落: %s\n", truncate(paragraphID, 8))
	},
}

var paragraphDeleteCmd = &cobra.Command{
	Use:   "delete <章节号> <段落ID>",
	Short: "删除段落",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		paragraphID := args[1]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确定要删除段落 %s 吗？\n", paragraphID)
			fmt.Print("输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		// 加载段落
		paragraphs, err := jsonStore.LoadChapterParagraphs(bookName, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并删除
		found := false
		var newParagraphs []model.Paragraph
		for _, p := range paragraphs.Paragraphs {
			if p.ID == paragraphID || truncate(p.ID, 8) == paragraphID {
				found = true
				continue
			}
			newParagraphs = append(newParagraphs, p)
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 段落 %s 不存在\n", paragraphID)
			return
		}

		paragraphs.Paragraphs = newParagraphs

		// 保存
		if err := jsonStore.SaveChapterParagraphs(bookName, paragraphs); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已删除段落: %s\n", truncate(paragraphID, 8))
	},
}

var paragraphMoveCmd = &cobra.Command{
	Use:   "move <章节号> <段落ID>",
	Short: "移动段落位置",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		paragraphID := args[1]
		position, _ := cmd.Flags().GetInt("position")

		// 移动段落
		if err := jsonStore.MoveParagraph(bookName, chapterID, paragraphID, position); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		fmt.Printf("✅ 已移动段落 %s 到位置 %d\n", truncate(paragraphID, 8), position)
	},
}

func init() {
	rootCmd.AddCommand(paragraphCmd)

	paragraphCmd.AddCommand(paragraphListCmd)
	paragraphCmd.AddCommand(paragraphShowCmd)
	paragraphCmd.AddCommand(paragraphAddCmd)
	paragraphCmd.AddCommand(paragraphEditCmd)
	paragraphCmd.AddCommand(paragraphDeleteCmd)
	paragraphCmd.AddCommand(paragraphMoveCmd)

	paragraphAddCmd.Flags().String("text", "", "段落内容")
	paragraphAddCmd.Flags().Int("position", -1, "插入位置（默认追加到末尾）")

	paragraphEditCmd.Flags().String("text", "", "段落内容")

	paragraphDeleteCmd.Flags().BoolP("force", "f", false, "强制删除，不确认")

	paragraphMoveCmd.Flags().IntP("position", "p", 0, "目标位置")
}