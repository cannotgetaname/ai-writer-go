package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"ai-writer/internal/service"
)

// toolCmd represents the tool command
var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "智能工具箱",
	Long: `AI 写作辅助工具集合。

子命令:
  name        生成名称（人名/功法/法宝/宗门/地点）
  character   生成角色设定
  conflict    生成冲突设计
  scene       生成场景描写
  goldfinger  生成金手指设定`,
}

var toolNameCmd = &cobra.Command{
	Use:   "name",
	Short: "生成名称",
	Long: `生成各类名称。

示例:
  ai-writer tool name --type 人名 --genre 玄幻 --count 5
  ai-writer tool name --type 功法 --genre 仙侠`,
	Run: func(cmd *cobra.Command, args []string) {
		nameType, _ := cmd.Flags().GetString("type")
		genre, _ := cmd.Flags().GetString("genre")
		count, _ := cmd.Flags().GetInt("count")
		gender, _ := cmd.Flags().GetString("gender")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		toolbox := service.NewToolboxService(llmClient)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := toolbox.GenerateNames(ctx, &service.NamingRequest{
			Type:   nameType,
			Genre:  genre,
			Count:  count,
			Gender: gender,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Printf("【%s名称生成结果】\n", nameType)
		fmt.Println("────────────────────────────────")
		for i, name := range result.Names {
			fmt.Printf("%d. %s\n", i+1, name.Name)
			if name.Meaning != "" {
				fmt.Printf("   寓意: %s\n", name.Meaning)
			}
		}
	},
}

var toolCharacterCmd = &cobra.Command{
	Use:   "character",
	Short: "生成角色设定",
	Long: `生成完整的角色设定。

示例:
  ai-writer tool character --type 主角 --gender 男 --genre 玄幻
  ai-writer tool character --type 反派 --gender 女 --theme 冷血无情`,
	Run: func(cmd *cobra.Command, args []string) {
		charType, _ := cmd.Flags().GetString("type")
		gender, _ := cmd.Flags().GetString("gender")
		genre, _ := cmd.Flags().GetString("genre")
		theme, _ := cmd.Flags().GetString("theme")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		toolbox := service.NewToolboxService(llmClient)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := toolbox.GenerateCharacter(ctx, &service.CharacterRequest{
			Type:   charType,
			Gender: gender,
			Genre:  genre,
			Theme:  theme,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Println("【角色设定】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("姓名: %s\n", result.Name)
		fmt.Printf("角色: %s\n", result.Role)
		fmt.Printf("性别: %s\n", result.Gender)
		if result.Personality != "" {
			fmt.Printf("性格: %s\n", result.Personality)
		}
		if result.Goal != "" {
			fmt.Printf("目标: %s\n", result.Goal)
		}
		if result.Background != "" {
			fmt.Printf("背景: %s\n", result.Background)
		}
		if result.Bio != "" {
			fmt.Printf("简介: %s\n", result.Bio)
		}
	},
}

var toolConflictCmd = &cobra.Command{
	Use:   "conflict",
	Short: "生成冲突设计",
	Long: `生成故事冲突设计。

示例:
  ai-writer tool conflict --type 人物 --genre 玄幻 --context "两个修士争夺宝物"`,
	Run: func(cmd *cobra.Command, args []string) {
		conflictType, _ := cmd.Flags().GetString("type")
		genre, _ := cmd.Flags().GetString("genre")
		contextStr, _ := cmd.Flags().GetString("context")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		toolbox := service.NewToolboxService(llmClient)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := toolbox.GenerateConflict(ctx, &service.ConflictRequest{
			Type:    conflictType,
			Genre:   genre,
			Context: contextStr,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Println("【冲突设计】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("标题: %s\n", result.Title)
		fmt.Printf("描述: %s\n", result.Description)
		fmt.Printf("利害关系: %s\n", result.Stakes)
		fmt.Printf("可能解决: %s\n", result.Resolution)
	},
}

var toolSceneCmd = &cobra.Command{
	Use:   "scene",
	Short: "生成场景描写",
	Long: `生成详细的场景描写。

示例:
  ai-writer tool scene --type 战斗 --location "紫山之巅" --characters "叶凡,姬紫月"`,
	Run: func(cmd *cobra.Command, args []string) {
		sceneType, _ := cmd.Flags().GetString("type")
		location, _ := cmd.Flags().GetString("location")
		characters, _ := cmd.Flags().GetString("characters")
		mood, _ := cmd.Flags().GetString("mood")
		description, _ := cmd.Flags().GetString("description")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		toolbox := service.NewToolboxService(llmClient)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := toolbox.GenerateScene(ctx, &service.SceneRequest{
			Type:        sceneType,
			Location:    location,
			Characters:  characters,
			Mood:        mood,
			Description: description,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Println("【场景描写】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("标题: %s\n", result.Title)
		fmt.Printf("环境: %s\n", result.Setting)
		fmt.Printf("氛围: %s\n", result.Atmosphere)
		fmt.Println()
		fmt.Println(result.Description)
	},
}

var toolGoldfingerCmd = &cobra.Command{
	Use:   "goldfinger",
	Short: "生成金手指设定",
	Long: `生成金手指/作弊器设定。

示例:
  ai-writer tool goldfinger --type 系统 --genre 都市 --theme "签到系统"`,
	Run: func(cmd *cobra.Command, args []string) {
		gfType, _ := cmd.Flags().GetString("type")
		genre, _ := cmd.Flags().GetString("genre")
		theme, _ := cmd.Flags().GetString("theme")
		level, _ := cmd.Flags().GetString("level")

		llmClient, err := initLLMClient()
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		toolbox := service.NewToolboxService(llmClient)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := toolbox.GenerateGoldfinger(ctx, &service.GoldfingerRequest{
			Type:  gfType,
			Genre: genre,
			Theme:  theme,
			Level:  level,
		})
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
			return
		}

		fmt.Println("【金手指设定】")
		fmt.Println("────────────────────────────────")
		fmt.Printf("名称: %s\n", result.Name)
		fmt.Printf("类型: %s\n", result.Type)
		fmt.Printf("描述: %s\n", result.Description)
		fmt.Printf("限制: %s\n", result.Limitations)
		fmt.Printf("来源: %s\n", result.Origin)
	},
}

func init() {
	rootCmd.AddCommand(toolCmd)

	toolCmd.AddCommand(toolNameCmd)
	toolCmd.AddCommand(toolCharacterCmd)
	toolCmd.AddCommand(toolConflictCmd)
	toolCmd.AddCommand(toolSceneCmd)
	toolCmd.AddCommand(toolGoldfingerCmd)

	// name 命令选项
	toolNameCmd.Flags().String("type", "人名", "名称类型 (人名/功法/法宝/宗门/地点)")
	toolNameCmd.Flags().String("genre", "玄幻", "题材风格")
	toolNameCmd.Flags().Int("count", 5, "生成数量")
	toolNameCmd.Flags().String("gender", "", "性别 (男/女)")

	// character 命令选项
	toolCharacterCmd.Flags().String("type", "主角", "角色类型 (主角/配角/反派)")
	toolCharacterCmd.Flags().String("gender", "男", "性别")
	toolCharacterCmd.Flags().String("genre", "玄幻", "题材")
	toolCharacterCmd.Flags().String("theme", "", "主题/特点")

	// conflict 命令选项
	toolConflictCmd.Flags().String("type", "人物", "冲突类型 (人物/利益/情感/理念)")
	toolConflictCmd.Flags().String("genre", "玄幻", "题材")
	toolConflictCmd.Flags().String("context", "", "背景上下文")

	// scene 命令选项
	toolSceneCmd.Flags().String("type", "日常", "场景类型 (战斗/日常/对话/冒险)")
	toolSceneCmd.Flags().String("location", "", "地点")
	toolSceneCmd.Flags().String("characters", "", "涉及角色")
	toolSceneCmd.Flags().String("mood", "", "氛围")
	toolSceneCmd.Flags().String("description", "", "简要描述")

	// goldfinger 命令选项
	toolGoldfingerCmd.Flags().String("type", "系统", "类型 (系统/天赋/宝物/传承)")
	toolGoldfingerCmd.Flags().String("genre", "玄幻", "题材")
	toolGoldfingerCmd.Flags().String("theme", "", "主题")
	toolGoldfingerCmd.Flags().String("level", "中等", "强度等级")

	// 书名生成命令
	var toolTitleCmd = &cobra.Command{
		Use:   "title",
		Short: "生成书名",
		Long: `生成吸引眼球的书名。

示例:
  ai-writer tool title --genre 玄幻 --count 5
  ai-writer tool title --genre 都市 --style 霸气`,
		Run: func(cmd *cobra.Command, args []string) {
			genre, _ := cmd.Flags().GetString("genre")
			theme, _ := cmd.Flags().GetString("theme")
			count, _ := cmd.Flags().GetInt("count")
			style, _ := cmd.Flags().GetString("style")

			llmClient, err := initLLMClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			toolbox := service.NewToolboxService(llmClient)
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result, err := toolbox.GenerateTitle(ctx, &service.TitleRequest{
				Genre: genre,
				Theme: theme,
				Count: count,
				Style: style,
			})
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			fmt.Println("【书名生成结果】")
			fmt.Println("────────────────────────────────")
			for i, t := range result.Titles {
				fmt.Printf("%d. %s\n", i+1, t.Title)
				if t.Meaning != "" {
					fmt.Printf("   寓意: %s\n", t.Meaning)
				}
				if t.Attraction != "" {
					fmt.Printf("   吸引力: %s\n", t.Attraction)
				}
			}
		},
	}

	// 简介生成命令
	var toolSynopsisCmd = &cobra.Command{
		Use:   "synopsis",
		Short: "生成简介",
		Long: `生成作品简介。

示例:
  ai-writer tool synopsis --genre 玄幻 --main-char "少年天才"
  ai-writer tool synopsis --genre 仙侠 --type long`,
		Run: func(cmd *cobra.Command, args []string) {
			genre, _ := cmd.Flags().GetString("genre")
			mainChar, _ := cmd.Flags().GetString("main-char")
			worldView, _ := cmd.Flags().GetString("world-view")
			synType, _ := cmd.Flags().GetString("type")

			llmClient, err := initLLMClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			toolbox := service.NewToolboxService(llmClient)
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result, err := toolbox.GenerateSynopsis(ctx, &service.SynopsisRequest{
				Genre:     genre,
				MainChar:  mainChar,
				WorldView: worldView,
				Type:      synType,
			})
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			fmt.Println("【作品简介】")
			fmt.Println("────────────────────────────────")
			fmt.Println(result.Synopsis)

			if len(result.Highlights) > 0 {
				fmt.Println("\n【卖点】")
				for _, h := range result.Highlights {
					fmt.Printf("  - %s\n", h)
				}
			}

			if result.Hook != "" {
				fmt.Println("\n【开篇钩子】")
				fmt.Println(result.Hook)
			}
		},
	}

	// 剧情转折命令
	var toolTwistCmd = &cobra.Command{
		Use:   "twist",
		Short: "生成剧情转折",
		Long: `生成意想不到的剧情转折。

示例:
  ai-writer tool twist --type unexpected --genre 玄幻
  ai-writer tool twist --type reversal --context "主角发现师父是杀父仇人"`,
		Run: func(cmd *cobra.Command, args []string) {
			twistType, _ := cmd.Flags().GetString("type")
			genre, _ := cmd.Flags().GetString("genre")
			contextStr, _ := cmd.Flags().GetString("context")
			characters, _ := cmd.Flags().GetString("characters")

			llmClient, err := initLLMClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			toolbox := service.NewToolboxService(llmClient)
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result, err := toolbox.GenerateTwist(ctx, &service.TwistRequest{
				Type:       twistType,
				Genre:      genre,
				Context:    contextStr,
				Characters: characters,
			})
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			fmt.Println("【剧情转折】")
			fmt.Println("────────────────────────────────")
			fmt.Printf("标题: %s\n", result.Title)
			fmt.Printf("描述: %s\n", result.Description)
			fmt.Printf("影响: %s\n", result.Impact)
			fmt.Printf("铺垫建议: %s\n", result.Setup)

			if len(result.Clues) > 0 {
				fmt.Println("\n【伏笔线索】")
				for _, c := range result.Clues {
					fmt.Printf("  - %s\n", c)
				}
			}
		},
	}

	// 对话生成命令
	var toolDialogueCmd = &cobra.Command{
		Use:   "dialogue",
		Short: "生成对话",
		Long: `生成角色对话。

示例:
  ai-writer tool dialogue --characters "叶凡,姬紫月" --situation "初次相遇" --mood 紧张`,
		Run: func(cmd *cobra.Command, args []string) {
			characters, _ := cmd.Flags().GetString("characters")
			situation, _ := cmd.Flags().GetString("situation")
			mood, _ := cmd.Flags().GetString("mood")
			genre, _ := cmd.Flags().GetString("genre")

			llmClient, err := initLLMClient()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			toolbox := service.NewToolboxService(llmClient)
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			result, err := toolbox.GenerateDialogue(ctx, &service.DialogueRequest{
				Characters: characters,
				Situation:  situation,
				Mood:       mood,
				Genre:      genre,
			})
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "错误: %v\n", err)
				return
			}

			fmt.Println("【对话内容】")
			fmt.Println("────────────────────────────────")
			fmt.Println(result.Content)
		},
	}

	toolCmd.AddCommand(toolTitleCmd)
	toolCmd.AddCommand(toolSynopsisCmd)
	toolCmd.AddCommand(toolTwistCmd)
	toolCmd.AddCommand(toolDialogueCmd)

	// title 命令选项
	toolTitleCmd.Flags().String("genre", "玄幻", "题材")
	toolTitleCmd.Flags().String("theme", "", "主题")
	toolTitleCmd.Flags().Int("count", 5, "数量")
	toolTitleCmd.Flags().String("style", "", "风格 (霸气/文艺/悬疑等)")

	// synopsis 命令选项
	toolSynopsisCmd.Flags().String("genre", "玄幻", "题材")
	toolSynopsisCmd.Flags().String("main-char", "", "主角设定")
	toolSynopsisCmd.Flags().String("world-view", "", "世界观")
	toolSynopsisCmd.Flags().String("type", "short", "类型 (short/long)")

	// twist 命令选项
	toolTwistCmd.Flags().String("type", "unexpected", "类型 (unexpected/reversal)")
	toolTwistCmd.Flags().String("genre", "玄幻", "题材")
	toolTwistCmd.Flags().String("context", "", "当前剧情上下文")
	toolTwistCmd.Flags().String("characters", "", "涉及角色")

	// dialogue 命令选项
	toolDialogueCmd.Flags().String("characters", "", "角色列表")
	toolDialogueCmd.Flags().String("situation", "", "场景情境")
	toolDialogueCmd.Flags().String("mood", "", "氛围")
	toolDialogueCmd.Flags().String("genre", "玄幻", "题材")
}