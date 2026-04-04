package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export <格式>",
	Short: "导出书籍",
	Long: `导出书籍内容到指定格式。

支持的格式:
  txt       纯文本格式
  markdown  Markdown 格式
  json      JSON 数据格式

示例:
  ai-writer export txt -o output.txt
  ai-writer export markdown -o output/
  ai-writer export txt --chapters 1-10 -o part1.txt`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format := args[0]

		name, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		outputFile, _ := cmd.Flags().GetString("output")
		chaptersRange, _ := cmd.Flags().GetString("chapters")

		// Load book
		book, err := jsonStore.LoadBook(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		switch format {
		case "txt":
			exportTXT(book, outputFile, chaptersRange)
		case "markdown", "md":
			exportMarkdown(book, outputFile, chaptersRange)
		case "json":
			exportJSON(book, outputFile)
		default:
			fmt.Fprintf(os.Stderr, "错误: 不支持的格式 '%s'\n", format)
			fmt.Println("支持的格式: txt, markdown, json")
		}
	},
}

func exportTXT(book *model.Book, outputFile, chaptersRange string) {
	chapters, err := jsonStore.LoadChapters(book.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	var content string
	content += fmt.Sprintf("%s\n\n", book.Name)
	content += "────────────────────────────────\n\n"

	// Parse chapter range
	start, end := parseRange(chaptersRange, len(chapters))

	for i := start; i <= end && i < len(chapters); i++ {
		ch := chapters[i]
		chContent, err := jsonStore.LoadChapterContent(book.Name, ch.ID)
		if err != nil {
			continue
		}

		content += fmt.Sprintf("第%d章 %s\n\n", ch.ID, ch.Title)
		content += chContent
		content += "\n\n────────────────────────────────\n\n"
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}
		fmt.Printf("✅ 已导出到: %s\n", outputFile)
	} else {
		fmt.Println(content)
	}
}

func exportMarkdown(book *model.Book, outputFile, chaptersRange string) {
	chapters, err := jsonStore.LoadChapters(book.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		return
	}

	var content string
	content += fmt.Sprintf("# %s\n\n", book.Name)
	content += fmt.Sprintf("> 创建时间: %s\n\n", book.CreatedAt.Format("2006-01-02"))

	// Parse chapter range
	start, end := parseRange(chaptersRange, len(chapters))

	for i := start; i <= end && i < len(chapters); i++ {
		ch := chapters[i]
		chContent, err := jsonStore.LoadChapterContent(book.Name, ch.ID)
		if err != nil {
			continue
		}

		content += fmt.Sprintf("## 第%d章 %s\n\n", ch.ID, ch.Title)
		content += chContent
		content += "\n\n"
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}
		fmt.Printf("✅ 已导出到: %s\n", outputFile)
	} else {
		fmt.Println(content)
	}
}

func exportJSON(book *model.Book, outputFile string) {
	// Load all data
	chapters, _ := jsonStore.LoadChapters(book.Name)
	characters, _ := jsonStore.LoadCharacters(book.Name)
	items, _ := jsonStore.LoadItems(book.Name)
	locations, _ := jsonStore.LoadLocations(book.Name)
	worldview, _ := jsonStore.LoadWorldView(book.Name)

	exportData := map[string]interface{}{
		"book":       book,
		"chapters":   chapters,
		"characters": characters,
		"items":      items,
		"locations":  locations,
		"worldview":  worldview,
	}

	// Load chapter contents
	chapterContents := make(map[int]string)
	for _, ch := range chapters {
		content, err := jsonStore.LoadChapterContent(book.Name, ch.ID)
		if err == nil {
			chapterContents[ch.ID] = content
		}
	}
	exportData["chapter_contents"] = chapterContents

	if outputFile != "" {
		printJSONToFile(exportData, outputFile)
		fmt.Printf("✅ 已导出到: %s\n", outputFile)
	} else {
		printJSON(exportData)
	}
}

func parseRange(rangeStr string, maxLen int) (start, end int) {
	start = 0
	end = maxLen - 1

	if rangeStr == "" {
		return start, end
	}

	// Parse "1-10" or "5" format
	var s, e int
	found := false
	for i, c := range rangeStr {
		if c == '-' {
			s = parseChapterID(rangeStr[:i])
			e = parseChapterID(rangeStr[i+1:])
			found = true
			break
		}
	}

	if found {
		if s > 0 {
			start = s - 1
		}
		if e > 0 && e <= maxLen {
			end = e - 1
		}
	} else {
		// Single chapter
		n := parseChapterID(rangeStr)
		if n > 0 && n <= maxLen {
			start = n - 1
			end = n - 1
		}
	}

	return start, end
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("output", "o", "", "输出文件路径")
	exportCmd.Flags().String("chapters", "", "导出章节范围 (如: 1-10)")
}