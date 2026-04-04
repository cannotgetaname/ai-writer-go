package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-writer/internal/llm"
)

// ArchitectService 架构师服务
type ArchitectService struct {
	llmClient llm.Client
}

// NewArchitectService 创建架构师服务
func NewArchitectService(llmClient llm.Client) *ArchitectService {
	return &ArchitectService{
		llmClient: llmClient,
	}
}

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusPlanned NodeStatus = "planned" // 规划中
	NodeStatusWriting NodeStatus = "writing" // 写作中
	NodeStatusDone    NodeStatus = "done"    // 已完成
	NodeStatusReview  NodeStatus = "review"  // 待审核
	NodeStatusHold    NodeStatus = "hold"    // 暂缓
)

// TreeNode 树节点
type TreeNode struct {
	ID          string     `json:"id"`
	Label       string     `json:"label"`
	Type        string     `json:"type"`        // root/volume/chapter
	Status      NodeStatus `json:"status"`
	Outline     string     `json:"outline"`
	Children    []*TreeNode `json:"children,omitempty"`
	ParentID    string     `json:"parent_id,omitempty"`
}

// FissionStrategy 分形裂变策略
type FissionStrategy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

// FissionRequest 分形裂变请求
type FissionRequest struct {
	BookName    string `json:"book_name"`
	NodeID      string `json:"node_id"`
	NodeType    string `json:"node_type"`    // volume/chapter
	CurrentOutline string `json:"current_outline"`
	Strategy    string `json:"strategy"`     // expand/refine/branch
	Count       int    `json:"count"`        // 生成数量
}

// FissionResult 分形裂变结果
type FissionResult struct {
	Nodes []*TreeNode `json:"nodes"`
}

// GenerateOutlineRequest 生成大纲请求
type GenerateOutlineRequest struct {
	BookName    string `json:"book_name"`
	Genre       string `json:"genre"`       // 题材
	MainChar    string `json:"main_char"`   // 主角设定
	Theme       string `json:"theme"`       // 主题
	TargetWords int    `json:"target_words"` // 目标字数
	VolumeCount int    `json:"volume_count"` // 分卷数量
}

// GenerateOutlineResult 生成大纲结果
type GenerateOutlineResult struct {
	Volumes []*TreeNode `json:"volumes"`
	Synopsis string      `json:"synopsis"` // 故事梗概
}

// GenerateOutline 生成大纲
func (s *ArchitectService) GenerateOutline(ctx context.Context, req *GenerateOutlineRequest) (*GenerateOutlineResult, error) {
	prompt := fmt.Sprintf(`请为一部%s题材的网络小说生成大纲框架。

主角设定: %s
主题: %s
目标字数: %d字
分卷数量: %d卷

请生成：
1. 故事梗概（200字以内）
2. 分卷大纲（每卷标题和主要内容）

请用JSON格式输出：
{
  "synopsis": "故事梗概",
  "volumes": [
    {
      "id": "vol_1",
      "label": "第一卷标题",
      "outline": "本卷主要内容概述",
      "children": [
        {"id": "chap_1", "label": "第1章标题", "outline": "章节大纲"}
      ]
    }
  ]
}`, req.Genre, req.MainChar, req.Theme, req.TargetWords, req.VolumeCount)

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	// 解析结果
	var outline GenerateOutlineResult
	if err := parseJSON(result, &outline); err != nil {
		// 尝试简单解析
		outline.Synopsis = extractJSONValue(result, "synopsis")
	}

	return &outline, nil
}

// Fission 分形裂变
func (s *ArchitectService) Fission(ctx context.Context, req *FissionRequest) (*FissionResult, error) {
	var prompt string

	switch req.Strategy {
	case "expand":
		prompt = fmt.Sprintf(`请将以下%s大纲展开为更详细的子节点。

当前大纲: %s
需要生成数量: %d

请生成详细的子节点，每个节点包含标题和概述。
用JSON格式输出：
{
  "nodes": [
    {"id": "node_1", "label": "标题", "outline": "概述", "status": "planned"}
  ]
}`, req.NodeType, req.CurrentOutline, req.Count)

	case "refine":
		prompt = fmt.Sprintf(`请优化以下%s大纲，使其更加完整和有吸引力。

当前大纲: %s

请优化并输出：
{
  "nodes": [
    {"id": "node_1", "label": "优化后标题", "outline": "优化后概述", "status": "planned"}
  ]
}`, req.NodeType, req.CurrentOutline)

	case "branch":
		prompt = fmt.Sprintf(`请基于以下%s大纲，生成多个可能的剧情分支。

当前大纲: %s
需要分支数量: %d

每个分支代表不同的剧情发展方向。
用JSON格式输出：
{
  "nodes": [
    {"id": "branch_1", "label": "分支标题", "outline": "分支概述", "status": "planned"}
  ]
}`, req.NodeType, req.CurrentOutline, req.Count)

	default:
		prompt = fmt.Sprintf(`请展开以下大纲: %s`, req.CurrentOutline)
	}

	result, err := s.llmClient.Call(ctx, prompt, "architect")
	if err != nil {
		return nil, err
	}

	var fissionResult FissionResult
	if err := parseJSON(result, &fissionResult); err != nil {
		fissionResult.Nodes = []*TreeNode{
			{
				ID:      fmt.Sprintf("node_%d", time.Now().Unix()),
				Label:   "生成结果",
				Outline: result,
				Status:  NodeStatusPlanned,
			},
		}
	}

	return &fissionResult, nil
}

// GetFissionStrategies 获取分形裂变策略
func GetFissionStrategies() map[string][]FissionStrategy {
	return map[string][]FissionStrategy{
		"expand": {
			{ID: "expand_detail", Name: "详细展开", Description: "将简单大纲展开为详细章节", Prompt: "展开为详细章节"},
			{ID: "expand_plot", Name: "剧情展开", Description: "展开剧情细节和转折", Prompt: "展开剧情细节"},
		},
		"refine": {
			{ID: "refine_logic", Name: "逻辑优化", Description: "优化剧情逻辑", Prompt: "优化逻辑"},
			{ID: "refine_pacing", Name: "节奏优化", Description: "优化叙事节奏", Prompt: "优化节奏"},
		},
		"branch": {
			{ID: "branch_plot", Name: "剧情分支", Description: "生成多条剧情线", Prompt: "生成剧情分支"},
			{ID: "branch_ending", Name: "结局分支", Description: "生成多种可能结局", Prompt: "生成结局分支"},
		},
	}
}

// AnalyzeStructure 分析结构
func (s *ArchitectService) AnalyzeStructure(ctx context.Context, bookName string, tree []*TreeNode) (map[string]interface{}, error) {
	// 统计节点状态
	stats := map[string]int{
		"planned": 0,
		"writing": 0,
		"done":    0,
		"review":  0,
		"hold":    0,
	}

	var countNodes func(nodes []*TreeNode)
	countNodes = func(nodes []*TreeNode) {
		for _, node := range nodes {
			stats[string(node.Status)]++
			if node.Children != nil {
				countNodes(node.Children)
			}
		}
	}
	countNodes(tree)

	total := 0
	for _, v := range stats {
		total += v
	}

	return map[string]interface{}{
		"total":   total,
		"stats":   stats,
		"progress": float64(stats["done"]+stats["review"]) / float64(total) * 100,
	}, nil
}

// parseJSON 解析JSON
func parseJSON(s string, v interface{}) error {
	// 提取JSON部分
	start := indexOf(s, "{")
	end := lastIndexOf(s, "}")
	if start == -1 || end == -1 {
		return fmt.Errorf("no JSON found")
	}
	jsonStr := s[start : end+1]
	return json.Unmarshal([]byte(jsonStr), v)
}

func indexOf(s string, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func lastIndexOf(s string, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}