package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// bookCmd represents the book command
var bookCmd = &cobra.Command{
	Use:   "book",
	Short: "书籍管理",
	Long: `管理小说项目，包括创建、删除、查看等操作。

子命令:
  list      列出所有书籍
  create    创建新书
  delete    删除书籍
  use       切换当前书籍
  info      查看书籍信息`,
}

var bookListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出所有书籍",
	Run: func(cmd *cobra.Command, args []string) {
		books, err := jsonStore.ListBooks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(books) == 0 {
			fmt.Println("暂无书籍，使用 'ai-writer book create <书名>' 创建新书")
			return
		}

		switch outputFormat {
		case "json":
			printJSON(books)
		case "markdown":
			fmt.Println("# 书籍列表")
			fmt.Println()
			for _, b := range books {
				fmt.Printf("- **%s** (创建于 %s)\n", b.Name, b.CreatedAt.Format("2006-01-02"))
			}
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "书名\t创建时间\t更新时间")
			fmt.Fprintln(w, "----\t--------\t--------")
			for _, b := range books {
				fmt.Fprintf(w, "%s\t%s\t%s\n",
					b.Name,
					b.CreatedAt.Format("2006-01-02"),
					b.UpdatedAt.Format("2006-01-02"),
				)
			}
			w.Flush()
		}
	},
}

var bookCreateCmd = &cobra.Command{
	Use:   "create <书名>",
	Short: "创建新书",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		book, err := jsonStore.CreateBook(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 设置为当前书籍
		setLastOpenBook(name)

		fmt.Printf("✅ 已创建书籍: %s\n", book.Name)
		if verbose {
			fmt.Printf("   ID: %s\n", book.ID)
			fmt.Printf("   创建时间: %s\n", book.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("\n已设置为当前书籍，后续命令无需指定 -b %s\n", name)
	},
}

var bookDeleteCmd = &cobra.Command{
	Use:   "delete <书名>",
	Short: "删除书籍",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确定要删除书籍 '%s' 吗？此操作不可恢复！\n", name)
			fmt.Print("输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		if err := jsonStore.DeleteBook(name); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		fmt.Printf("✅ 已删除书籍: %s\n", name)
	},
}

var bookInfoCmd = &cobra.Command{
	Use:   "info [书名]",
	Short: "查看书籍详情",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
			// 设置为当前书籍
			setLastOpenBook(name)
		} else {
			var err error
			name, err = requireBookName()
			if err != nil {
				fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				return
			}
		}

		// 加载书籍信息
		volumes, _ := jsonStore.LoadVolumes(name)
		chapters, _ := jsonStore.LoadChapters(name)
		characters, _ := jsonStore.LoadCharacters(name)
		items, _ := jsonStore.LoadItems(name)
		locations, _ := jsonStore.LoadLocations(name)

		fmt.Printf("📖 %s\n", name)
		fmt.Println("────────────────────────────────")
		fmt.Printf("  分卷数: %d\n", len(volumes))
		fmt.Printf("  章节数: %d\n", len(chapters))
		fmt.Printf("  人物数: %d\n", len(characters))
		fmt.Printf("  物品数: %d\n", len(items))
		fmt.Printf("  地点数: %d\n", len(locations))

		// 统计字数
		var totalWords int
		for _, ch := range chapters {
			content, err := jsonStore.LoadChapterContent(name, ch.ID)
			if err == nil {
				totalWords += countChineseChars(content)
			}
		}
		fmt.Printf("  总字数: %d\n", totalWords)
	},
}

var bookUseCmd = &cobra.Command{
	Use:   "use <书名>",
	Short: "切换当前书籍",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// 验证书籍是否存在
		books, err := jsonStore.ListBooks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		found := false
		for _, b := range books {
			if b.Name == name {
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 书籍 '%s' 不存在\n", name)
			fmt.Println("使用 'ai-writer book list' 查看所有书籍")
			return
		}

		// 设置为当前书籍
		setLastOpenBook(name)

		fmt.Printf("✅ 已切换到书籍: %s\n", name)
		fmt.Println("后续命令无需指定 -b 参数")
	},
}

func init() {
	rootCmd.AddCommand(bookCmd)

	bookCmd.AddCommand(bookListCmd)
	bookCmd.AddCommand(bookCreateCmd)
	bookCmd.AddCommand(bookDeleteCmd)
	bookCmd.AddCommand(bookInfoCmd)
	bookCmd.AddCommand(bookUseCmd)

	bookDeleteCmd.Flags().BoolP("force", "f", false, "强制删除，不确认")
}

// countChineseChars counts Chinese characters in a string
func countChineseChars(s string) int {
	count := 0
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			count++
		}
	}
	return count
}