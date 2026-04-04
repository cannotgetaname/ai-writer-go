package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"ai-writer/internal/model"
)

// itemCmd represents the item command
var itemCmd = &cobra.Command{
	Use:   "item",
	Short: "物品管理",
	Long: `管理书籍物品，包括查看、添加、编辑、删除等操作。

子命令:
  list      列出所有物品
  add       添加物品
  show      查看物品详情
  edit      编辑物品
  delete    删除物品`,
}

var itemListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出物品",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		items, err := jsonStore.LoadItems(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		if len(items) == 0 {
			fmt.Println("暂无物品，使用 'ai-writer item add' 创建物品")
			return
		}

		switch outputFormat {
		case "json":
			printJSON(items)
		case "markdown":
			fmt.Println("# 物品列表")
			fmt.Println()
			for _, item := range items {
				fmt.Printf("- **%s** (%s) - %s\n", item.Name, item.Type, item.Owner)
			}
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "名称\t类型\t持有者\t简介")
			fmt.Fprintln(w, "----\t----\t------\t----")
			for _, item := range items {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					item.Name, item.Type, item.Owner, truncate(item.Description, 30))
			}
			w.Flush()
		}
	},
}

var itemShowCmd = &cobra.Command{
	Use:   "show <物品名>",
	Short: "查看物品详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		itemName := args[0]
		items, err := jsonStore.LoadItems(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找物品
		var item *model.Item
		for _, i := range items {
			if i.Name == itemName {
				item = i
				break
			}
		}

		if item == nil {
			fmt.Fprintf(os.Stderr, "错误: 物品 '%s' 不存在\n", itemName)
			return
		}

		fmt.Printf("📦 %s\n", item.Name)
		fmt.Println("────────────────────────────────")
		fmt.Printf("  类型: %s\n", item.Type)
		fmt.Printf("  持有者: %s\n", item.Owner)
		fmt.Printf("  描述: %s\n", item.Description)
		if item.Origin != "" {
			fmt.Printf("  来历: %s\n", item.Origin)
		}
		if item.Abilities != "" {
			fmt.Printf("  能力: %s\n", item.Abilities)
		}
	},
}

var itemAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加物品",
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		itemType, _ := cmd.Flags().GetString("type")
		owner, _ := cmd.Flags().GetString("owner")
		description, _ := cmd.Flags().GetString("description")
		origin, _ := cmd.Flags().GetString("origin")
		abilities, _ := cmd.Flags().GetString("abilities")

		if name == "" {
			fmt.Fprintf(os.Stderr, "错误: 请指定物品名称 (--name)\n")
			return
		}

		// 加载现有物品
		items, err := jsonStore.LoadItems(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 检查是否已存在
		for _, i := range items {
			if i.Name == name {
				fmt.Fprintf(os.Stderr, "错误: 物品 '%s' 已存在\n", name)
				return
			}
		}

		// 创建新物品
		newItem := &model.Item{
			ID:          generateID(),
			BookID:      bookName,
			Name:        name,
			Type:        itemType,
			Owner:       owner,
			Description: description,
			Origin:      origin,
			Abilities:   abilities,
		}

		items = append(items, newItem)

		// 保存
		if err := jsonStore.SaveItems(bookName, items); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已添加物品: %s\n", name)
		if verbose {
			fmt.Printf("   类型: %s\n", itemType)
			fmt.Printf("   持有者: %s\n", owner)
			fmt.Printf("   描述: %s\n", description)
		}
	},
}

var itemEditCmd = &cobra.Command{
	Use:   "edit <物品名>",
	Short: "编辑物品",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		itemName := args[0]

		// 加载物品列表
		items, err := jsonStore.LoadItems(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找物品
		var item *model.Item
		for _, i := range items {
			if i.Name == itemName {
				item = i
				break
			}
		}

		if item == nil {
			fmt.Fprintf(os.Stderr, "错误: 物品 '%s' 不存在\n", itemName)
			return
		}

		// 获取编辑参数
		newName, _ := cmd.Flags().GetString("name")
		itemType, _ := cmd.Flags().GetString("type")
		owner, _ := cmd.Flags().GetString("owner")
		description, _ := cmd.Flags().GetString("description")

		// 更新字段
		if newName != "" {
			item.Name = newName
		}
		if itemType != "" {
			item.Type = itemType
		}
		if owner != "" {
			item.Owner = owner
		}
		if description != "" {
			item.Description = description
		}

		// 保存
		if err := jsonStore.SaveItems(bookName, items); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已更新物品: %s\n", item.Name)
		if verbose {
			fmt.Printf("   类型: %s\n", item.Type)
			fmt.Printf("   持有者: %s\n", item.Owner)
			fmt.Printf("   描述: %s\n", item.Description)
		}
	},
}

var itemDeleteCmd = &cobra.Command{
	Use:   "delete <物品名>",
	Short: "删除物品",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bookName, err := requireBookName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		itemName := args[0]

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("确定要删除物品 '%s' 吗？\n", itemName)
			fmt.Print("输入 'yes' 确认: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("已取消")
				return
			}
		}

		// 加载物品列表
		items, err := jsonStore.LoadItems(bookName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		// 查找并删除
		found := false
		var newItems []*model.Item
		for _, i := range items {
			if i.Name == itemName {
				found = true
				continue
			}
			newItems = append(newItems, i)
		}

		if !found {
			fmt.Fprintf(os.Stderr, "错误: 物品 '%s' 不存在\n", itemName)
			return
		}

		// 保存
		if err := jsonStore.SaveItems(bookName, newItems); err != nil {
			fmt.Fprintf(os.Stderr, "错误: 保存失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 已删除物品: %s\n", itemName)
	},
}

func init() {
	rootCmd.AddCommand(itemCmd)

	itemCmd.AddCommand(itemListCmd)
	itemCmd.AddCommand(itemShowCmd)
	itemCmd.AddCommand(itemAddCmd)
	itemCmd.AddCommand(itemEditCmd)
	itemCmd.AddCommand(itemDeleteCmd)

	itemAddCmd.Flags().String("name", "", "物品名称")
	itemAddCmd.Flags().String("type", "法宝", "物品类型 (武器/法宝/丹药/材料)")
	itemAddCmd.Flags().String("owner", "", "持有者")
	itemAddCmd.Flags().String("description", "", "描述")
	itemAddCmd.Flags().String("origin", "", "来历")
	itemAddCmd.Flags().String("abilities", "", "能力")

	itemEditCmd.Flags().String("name", "", "新名称")
	itemEditCmd.Flags().String("type", "", "物品类型")
	itemEditCmd.Flags().String("owner", "", "持有者")
	itemEditCmd.Flags().String("description", "", "描述")

	itemDeleteCmd.Flags().BoolP("force", "f", false, "强制删除，不确认")
}