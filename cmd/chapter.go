package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// chapterCmd represents the chapter command
var chapterCmd = &cobra.Command{
	Use:   "chapter",
	Short: "章节管理",
	Long: `管理书籍章节，包括查看、编辑、删除等操作。

子命令:
  list      列出所有章节
  show      查看章节内容
  edit      编辑章节（打开编辑器）
  delete    删除章节`,
}

var chapterListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出章节",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapters, err := jsonStore.LoadChapters(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(chapters) == 0 {
			fmt.Println("暂无章节")
			return
		}

		switch outputFormat {
		case "json":
			printJSON(chapters)
		case "markdown":
			fmt.Println("# 章节列表")
			fmt.Println()
			for _, ch := range chapters {
				fmt.Printf("- 第%d章: %s\n", ch.ID, ch.Title)
			}
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "章节\t标题\t大纲\t字数")
			fmt.Fprintln(w, "----\t----\t----\t----")
			for _, ch := range chapters {
				wordCount := 0
				content, err := jsonStore.LoadChapterContent(name, ch.ID)
				if err == nil {
					wordCount = countChineseChars(content)
				}
				fmt.Fprintf(w, "第%d章\t%s\t%s\t%d\n",
					ch.ID, ch.Title, truncate(ch.Outline, 30), wordCount)
			}
			w.Flush()
		}
	},
}

var chapterShowCmd = &cobra.Command{
	Use:   "show <章节号>",
	Short: "查看章节内容",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])
		chapters, err := jsonStore.LoadChapters(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// Find chapter
		var chapter *model.Chapter
		for _, ch := range chapters {
			if ch.ID == chapterID {
				chapter = ch
				break
			}
		}

		if chapter == nil {
			fmt.Fprintf(os.Stderr, "错误: 章节 %d 不存在\n", chapterID)
			return
		}

		// Load content
		content, err := jsonStore.LoadChapterContent(name, chapterID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if verbose {
			fmt.Printf("📖 第%d章: %s\n", chapter.ID, chapter.Title)
			fmt.Println("────────────────────────────────")
			fmt.Printf("大纲: %s\n", chapter.Outline)
			fmt.Printf("字数: %d\n", countChineseChars(content))
			fmt.Println("────────────────────────────────")
			fmt.Println()
		}

		fmt.Println(content)
	},
}

var chapterDeleteCmd = &cobra.Command{
	Use:   "delete <章节号>",
	Short: "删除章节",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确定要删除第%d章吗？此操作不可恢复！\n", chapterID)
			fmt.Print("输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		// 加载章节列表
		chapters, err := jsonStore.LoadChapters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并删除章节
		found := false
		var newChapters []*model.Chapter
		for _, ch := range chapters {
			if ch.ID == chapterID {
				found = true
				continue
			}
			newChapters = append(newChapters, ch)
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 章节 %d 不存在\n", chapterID)
			return
		}

		// 保存章节结构
		if err := jsonStore.SaveChapters(bookName, newChapters); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		// 删除段落文件
		if err := jsonStore.DeleteChapterParagraphs(bookName, chapterID); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 删除段落文件失败: %v\n", err)
		}

		fmt.Printf("✅ 已删除第%d章\n", chapterID)
	},
}

var chapterAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加章节",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		title, _ := cmd.Flags().GetString("title")
		outline, _ := cmd.Flags().GetString("outline")
		volumeID, _ := cmd.Flags().GetString("volume")

		if title == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定章节标题 (--title)\n")
			return
		}

		// 加载现有章节
		chapters, err := jsonStore.LoadChapters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 计算新章节 ID
		maxID := 0
		for _, ch := range chapters {
			if ch.ID > maxID {
				maxID = ch.ID
			}
		}
		newID := maxID + 1

		// 默认分卷
		if volumeID == "" {
			volumeID = "vol_1"
		}

		// 创建新章节
		now := time.Now()
		newChapter := &model.Chapter{
			ID:        newID,
			BookID:    bookName,
			VolumeID:  volumeID,
			Title:     title,
			Outline:   outline,
			TimeInfo:  model.TimeInfo{Label: "", Duration: "0", Events: []string{}},
			CreatedAt: now,
			UpdatedAt: now,
		}

		chapters = append(chapters, newChapter)

		// 保存
		if err := jsonStore.SaveChapters(bookName, chapters); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已添加第%d章: %s\n", newID, title)
	},
}

var chapterEditCmd = &cobra.Command{
	Use:   "edit <章节号>",
	Short: "编辑章节",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		chapterID := parseChapterID(args[0])

		// 加载章节列表
		chapters, err := jsonStore.LoadChapters(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找章节
		var chapter *model.Chapter
		for _, ch := range chapters {
			if ch.ID == chapterID {
				chapter = ch
				break
			}
		}

		if chapter == nil {
			fmt.Fprintf(os.Stderr, "错误: 章节 %d 不存在\n", chapterID)
			return
		}

		// 获取编辑参数
		title, _ := cmd.Flags().GetString("title")
		outline, _ := cmd.Flags().GetString("outline")

		// 更新字段
		if title != "" {
			chapter.Title = title
		}
		if outline != "" {
			chapter.Outline = outline
		}
		chapter.UpdatedAt = time.Now()

		// 保存
		if err := jsonStore.SaveChapters(bookName, chapters); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已更新第%d章\n", chapterID)
		if verbose {
			fmt.Printf("   标题: %s\n", chapter.Title)
			fmt.Printf("   大纲: %s\n", chapter.Outline)
		}
	},
}

func init() {
	rootCmd.AddCommand(chapterCmd)

	chapterCmd.AddCommand(chapterListCmd)
	chapterCmd.AddCommand(chapterShowCmd)
	chapterCmd.AddCommand(chapterDeleteCmd)
	chapterCmd.AddCommand(chapterAddCmd)
	chapterCmd.AddCommand(chapterEditCmd)

	chapterDeleteCmd.Flags().BoolP("force", "f", false, "强制删除，不确认")

	chapterAddCmd.Flags().String("title", "", "章节标题")
	chapterAddCmd.Flags().String("outline", "", "章节大纲")
	chapterAddCmd.Flags().String("volume", "", "所属分卷ID")

	chapterEditCmd.Flags().String("title", "", "章节标题")
	chapterEditCmd.Flags().String("outline", "", "章节大纲")
}

// parseChapterID parses chapter ID from string like "1" or "第1章"
func parseChapterID(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}