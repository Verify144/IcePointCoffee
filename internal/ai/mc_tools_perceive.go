package ai

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Verify144/IcePointCoffee/internal/mc"
)

// MCPerceiveTool 环境感知
type MCPerceiveTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCPerceiveTool() *MCPerceiveTool { return &MCPerceiveTool{} }
func (t *MCPerceiveTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCPerceiveTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCPerceiveTool) Name() string { return "mc_perceive" }
func (t *MCPerceiveTool) Description() string {
	return "感知周围环境（玩家/生物/方块），返回结构化数据供 AI 理解世界状态。" +
		"建议在执行建造指令前先用此工具确认当前位置和周围环境。"
}
func (t *MCPerceiveTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"scope": map[string]interface{}{
				"type": "string",
				"enum":  []string{"nearby", "players", "entities", "blocks", "all"},
				"description": "感知范围：nearby=附近一切, players=仅玩家, entities=仅生物, blocks=仅方块, all=全部",
				"default":    "nearby",
			},
			"radius": map[string]interface{}{
				"type":        "integer",
				"description": "扫描半径（方块），建议 5-30",
				"default":     10,
			},
			"detail": map[string]interface{}{
				"type": "string",
				"enum":  []string{"low", "medium", "high"},
				"description": "详细程度：low=快速, medium=标准, high=详细",
				"default":     "medium",
			},
		},
		"required": []string{},
	}
}
func (t *MCPerceiveTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Scope  string `json:"scope"`
		Radius int    `json:"radius"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Radius <= 0 {
		params.Radius = 10
	}
	if params.Radius > 50 {
		params.Radius = 50 // 安全上限
	}
	if params.Detail == "" {
		params.Detail = "medium"
	}
	if params.Scope == "" {
		params.Scope = "nearby"
	}

	client := t.getClient()
	if client == nil {
		return map[string]interface{}{
			"success":     false,
			"error":       "MC 客户端未初始化",
			"description": "抱歉，无法连接服务器来感知环境。",
		}, nil
	}

	// 过滤 scope
	if params.Scope == "players" {
		// 仅玩家感知
		ws := mc.NewWorldState()
		result := ws.Perceive(ctx, client, params.Radius, params.Detail)
		result.NearbyEntities = nil
		result.NearbyBlocks = nil
		if result.Description != "" {
			result.Description = "【玩家感知】" + result.Description
		}
		return result, nil
	}

	result, err := client.Perceive(params.Radius, params.Detail)
	if err != nil {
		return map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"description": "感知环境时出错：" + err.Error(),
		}, nil
	}

	// 按 scope 过滤
	switch params.Scope {
	case "entities":
		result.NearbyPlayers = nil
		result.NearbyBlocks = nil
		result.Description = "【生物感知】" + result.Description
	case "blocks":
		result.NearbyPlayers = nil
		result.NearbyEntities = nil
		result.Description = "【方块感知】" + result.Description
	case "nearby":
		// 全部但截断方块列表
		if len(result.NearbyBlocks) > 20 {
			result.NearbyBlocks = result.NearbyBlocks[:20]
		}
	}

	return result, nil
}

var _ Tool = (*MCPerceiveTool)(nil)
