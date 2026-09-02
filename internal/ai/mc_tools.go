package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/mc"
)

// MCCommandTool 执行任意 MC 命令（带白名单）
type MCCommandTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

// NewMCCommandTool 创建命令工具
func NewMCCommandTool() *MCCommandTool {
	return &MCCommandTool{}
}

// SetClient 注入 MC 客户端
func (t *MCCommandTool) SetClient(client mc.ClientInterface) {
	t.mu.Lock()
	t.client = client
	t.mu.Unlock()
}

func (t *MCCommandTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}

func (t *MCCommandTool) Name() string { return "mc_command" }
func (t *MCCommandTool) Description() string {
	return "执行任意 MC 命令（带白名单与危险检查）。命令不需要 / 前缀。可用于 setblock/fill/give/time/weather/gamemode 等。"
}

func (t *MCCommandTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "要执行的命令，例如 'time set day'",
			},
		},
		"required": []string{"command"},
	}
}

func (t *MCCommandTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Command == "" {
		return nil, fmt.Errorf("command 不能为空")
	}

	cmd := strings.TrimPrefix(strings.TrimSpace(params.Command), "/")

	// 黑名单
	if blocked := IsBlacklisted(cmd); blocked != "" {
		return map[string]interface{}{
			"success": false,
			"error":   "危险命令被拦截: " + blocked,
			"hint":    "如确需执行，请通过 Dashboard 手动开启确认模式",
		}, nil
	}

	client := t.getClient()
	if client == nil {
		return nil, fmt.Errorf("MC 客户端未初始化")
	}
	if !client.IsConnected() {
		return nil, fmt.Errorf("MC 客户端未连接服务器")
	}

	output, err := client.SendCommand(ctx, cmd)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"command": cmd,
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"command": cmd,
		"output":  output,
	}, nil
}

// MCChatTool 发送聊天消息
type MCChatTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCChatTool() *MCChatTool { return &MCChatTool{} }

func (t *MCChatTool) SetClient(client mc.ClientInterface) {
	t.mu.Lock()
	t.client = client
	t.mu.Unlock()
}

func (t *MCChatTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}

func (t *MCChatTool) Name() string { return "mc_chat" }
func (t *MCChatTool) Description() string {
	return "以机器人身份向服务器发送聊天消息"
}
func (t *MCChatTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{"type": "string", "description": "要发送的消息"},
		},
		"required": []string{"message"},
	}
}
func (t *MCChatTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SendChat(params.Message); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "sent": params.Message}, nil
}

// MCTeleportTool 传送玩家
type MCTeleportTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCTeleportTool() *MCTeleportTool { return &MCTeleportTool{} }
func (t *MCTeleportTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCTeleportTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCTeleportTool) Name() string { return "mc_teleport" }
func (t *MCTeleportTool) Description() string {
	return "传送玩家到指定坐标（可以是玩家或坐标 x,y,z）"
}
func (t *MCTeleportTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{"type": "string", "description": "玩家名（@s/@a/Steve）"},
			"x":      map[string]interface{}{"type": "number"},
			"y":      map[string]interface{}{"type": "number"},
			"z":      map[string]interface{}{"type": "number"},
		},
		"required": []string{"target", "x", "y", "z"},
	}
}
func (t *MCTeleportTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target string  `json:"target"`
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		Z      float64 `json:"z"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().Teleport(params.Target, params.X, params.Y, params.Z); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "teleported": params.Target, "to": map[string]float64{"x": params.X, "y": params.Y, "z": params.Z}}, nil
}

// MCGiveTool 给予物品
type MCGiveTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCGiveTool() *MCGiveTool { return &MCGiveTool{} }
func (t *MCGiveTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCGiveTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCGiveTool) Name() string { return "mc_give" }
func (t *MCGiveTool) Description() string {
	return "给予玩家物品（mc 物品 ID）"
}
func (t *MCGiveTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{"type": "string"},
			"item":   map[string]interface{}{"type": "string", "description": "物品 ID，如 'diamond'"},
			"count":  map[string]interface{}{"type": "integer", "default": 1},
		},
		"required": []string{"target", "item"},
	}
}
func (t *MCGiveTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target string `json:"target"`
		Item   string `json:"item"`
		Count  int    `json:"count"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().GiveItem(params.Target, params.Item, params.Count); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "gave": params.Item, "count": params.Count, "to": params.Target}, nil
}

// MCSetBlockTool 设置方块
type MCSetBlockTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCSetBlockTool() *MCSetBlockTool { return &MCSetBlockTool{} }
func (t *MCSetBlockTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCSetBlockTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCSetBlockTool) Name() string { return "mc_setblock" }
func (t *MCSetBlockTool) Description() string { return "在指定坐标放置方块" }
func (t *MCSetBlockTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"x":     map[string]interface{}{"type": "integer"},
			"y":     map[string]interface{}{"type": "integer"},
			"z":     map[string]interface{}{"type": "integer"},
			"block": map[string]interface{}{"type": "string", "description": "方块 ID，如 'diamond_block'"},
		},
		"required": []string{"x", "y", "z", "block"},
	}
}
func (t *MCSetBlockTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		X     int    `json:"x"`
		Y     int    `json:"y"`
		Z     int    `json:"z"`
		Block string `json:"block"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SetBlock(params.X, params.Y, params.Z, params.Block); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "placed": params.Block, "at": map[string]int{"x": params.X, "y": params.Y, "z": params.Z}}, nil
}

// MCFillTool 填充区域
type MCFillTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCFillTool() *MCFillTool { return &MCFillTool{} }
func (t *MCFillTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCFillTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCFillTool) Name() string { return "mc_fill" }
func (t *MCFillTool) Description() string {
	return "用指定方块填充一个长方体区域。⚠️ 大量方块请慎用，可能造成卡顿。"
}
func (t *MCFillTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"x1":    map[string]interface{}{"type": "integer"},
			"y1":    map[string]interface{}{"type": "integer"},
			"z1":    map[string]interface{}{"type": "integer"},
			"x2":    map[string]interface{}{"type": "integer"},
			"y2":    map[string]interface{}{"type": "integer"},
			"z2":    map[string]interface{}{"type": "integer"},
			"block": map[string]interface{}{"type": "string"},
		},
		"required": []string{"x1", "y1", "z1", "x2", "y2", "z2", "block"},
	}
}
func (t *MCFillTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		X1, Y1, Z1, X2, Y2, Z2 int
		Block                  string
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	// 体积检查
	volume := (params.X2 - params.X1 + 1) * (params.Y2 - params.Y1 + 1) * (params.Z2 - params.Z1 + 1)
	if volume > 100000 {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("体积过大: %d 方块（限制 100000）", volume),
		}, nil
	}

	if err := t.getClient().Fill(params.X1, params.Y1, params.Z1, params.X2, params.Y2, params.Z2, params.Block); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success": true,
		"filled":  params.Block,
		"volume":  volume,
	}, nil
}

// MCDialogTool 发送 tellraw/title 给指定玩家
type MCDialogTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCDialogTool() *MCDialogTool { return &MCDialogTool{} }
func (t *MCDialogTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCDialogTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCDialogTool) Name() string { return "mc_dialog" }
func (t *MCDialogTool) Description() string {
	return "向指定玩家发送富文本/标题（@s/@a/玩家名）"
}
func (t *MCDialogTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target":   map[string]interface{}{"type": "string"},
			"kind":     map[string]interface{}{"type": "string", "enum": []string{"tellraw", "title", "actionbar"}},
			"message":  map[string]interface{}{"type": "string"},
			"subtitle": map[string]interface{}{"type": "string", "description": "仅 kind=title 时使用"},
		},
		"required": []string{"target", "kind", "message"},
	}
}
func (t *MCDialogTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target   string `json:"target"`
		Kind     string `json:"kind"`
		Message  string `json:"message"`
		Subtitle string `json:"subtitle"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	client := t.getClient()
	switch params.Kind {
	case "tellraw":
		if err := client.SendTellRaw(params.Target, params.Message); err != nil {
			return map[string]interface{}{"success": false, "error": err.Error()}, nil
		}
	case "title", "actionbar":
		if params.Kind == "actionbar" {
			return nil, fmt.Errorf("actionbar 暂未支持")
		}
		if err := client.SendTitle(params.Target, params.Message, params.Subtitle); err != nil {
			return map[string]interface{}{"success": false, "error": err.Error()}, nil
		}
	default:
		return nil, fmt.Errorf("未知的 kind: %s", params.Kind)
	}
	return map[string]interface{}{"success": true, "kind": params.Kind, "target": params.Target}, nil
}

// MCGameModeTool 切换游戏模式
type MCGameModeTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCGameModeTool() *MCGameModeTool { return &MCGameModeTool{} }
func (t *MCGameModeTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCGameModeTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCGameModeTool) Name() string { return "mc_gamemode" }
func (t *MCGameModeTool) Description() string {
	return "切换玩家游戏模式（survival/creative/adventure/spectator）"
}
func (t *MCGameModeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{"type": "string"},
			"mode":   map[string]interface{}{"type": "string", "enum": []string{"survival", "creative", "adventure", "spectator"}},
		},
		"required": []string{"target", "mode"},
	}
}
func (t *MCGameModeTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target string `json:"target"`
		Mode   string `json:"mode"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SetGameMode(params.Target, params.Mode); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "target": params.Target, "mode": params.Mode}, nil
}

// MCWorldTool 世界设定（时间/天气）
type MCWorldTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCWorldTool() *MCWorldTool { return &MCWorldTool{} }
func (t *MCWorldTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCWorldTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCWorldTool) Name() string { return "mc_world" }
func (t *MCWorldTool) Description() string {
	return "设置世界时间或天气"
}
func (t *MCWorldTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"kind":    map[string]interface{}{"type": "string", "enum": []string{"time", "weather"}},
			"value":   map[string]interface{}{"type": "string", "description": "time: day/night/0..24000; weather: clear/rain/thunder"},
		},
		"required": []string{"kind", "value"},
	}
}
func (t *MCWorldTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Kind  string `json:"kind"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	client := t.getClient()
	switch params.Kind {
	case "time":
		if err := client.SetTime(params.Value); err != nil {
			return map[string]interface{}{"success": false, "error": err.Error()}, nil
		}
	case "weather":
		if err := client.SetWeather(params.Value); err != nil {
			return map[string]interface{}{"success": false, "error": err.Error()}, nil
		}
	default:
		return nil, fmt.Errorf("未知的 kind: %s", params.Kind)
	}
	return map[string]interface{}{"success": true, "kind": params.Kind, "value": params.Value}, nil
}

// MCStatusTool 查询状态
type MCStatusTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCStatusTool() *MCStatusTool { return &MCStatusTool{} }
func (t *MCStatusTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock()
	t.client = c
	t.mu.Unlock()
}
func (t *MCStatusTool) getClient() mc.ClientInterface {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
}
func (t *MCStatusTool) Name() string { return "mc_status" }
func (t *MCStatusTool) Description() string {
	return "查询当前 MC 连接状态"
}
func (t *MCStatusTool) Parameters() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}
func (t *MCStatusTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	c := t.getClient()
	if c == nil {
		return map[string]interface{}{"connected": false, "reason": "client not initialized"}, nil
	}
	return c.Status(), nil
}

// ==== 安全：黑名单 ====

// IsBlacklisted 返回被拦截的原因，空字符串表示允许
func IsBlacklisted(cmd string) string {
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))
	blacklist := []string{
		"stop", "kick", "ban", "banlist", "op", "deop", "pardon",
		"whitelist", "reload", "restart", "shutdown", "save-all off",
		"forcehost", "publish",
	}
	for _, b := range blacklist {
		// 检查命令是否以黑名单开头（处理 stopxxx 也算 stop）
		if strings.HasPrefix(cmdLower, b+" ") || cmdLower == b {
			return b
		}
	}
	return ""
}

// ==== 工具集合 ====

// RegisterMCTools 注册所有 MC 工具到 registry
// 返回一个 *MCController，用于动态注入/移除 client
func RegisterMCTools(reg *ToolRegistry) *MCController {
	command := NewMCCommandTool()
	chat := NewMCChatTool()
	tp := NewMCTeleportTool()
	give := NewMCGiveTool()
	setblock := NewMCSetBlockTool()
	fill := NewMCFillTool()
	dialog := NewMCDialogTool()
	gm := NewMCGameModeTool()
	world := NewMCWorldTool()
	status := NewMCStatusTool()

	reg.Register(command)
	reg.Register(chat)
	reg.Register(tp)
	reg.Register(give)
	reg.Register(setblock)
	reg.Register(fill)
	reg.Register(dialog)
	reg.Register(gm)
	reg.Register(world)
	reg.Register(status)

	return &MCController{
		Command:  command,
		Chat:     chat,
		Teleport: tp,
		Give:     give,
		SetBlock: setblock,
		Fill:     fill,
		Dialog:   dialog,
		GameMode: gm,
		World:    world,
		Status:   status,
	}
}

// MCController 持有所有 MC 工具，提供统一的 client 注入
type MCController struct {
	Command  *MCCommandTool
	Chat     *MCChatTool
	Teleport *MCTeleportTool
	Give     *MCGiveTool
	SetBlock *MCSetBlockTool
	Fill     *MCFillTool
	Dialog   *MCDialogTool
	GameMode *MCGameModeTool
	World    *MCWorldTool
	Status   *MCStatusTool
}

// Inject 注入 MC 客户端到所有工具
func (c *MCController) Inject(client mc.ClientInterface) {
	c.Command.SetClient(client)
	c.Chat.SetClient(client)
	c.Teleport.SetClient(client)
	c.Give.SetClient(client)
	c.SetBlock.SetClient(client)
	c.Fill.SetClient(client)
	c.Dialog.SetClient(client)
	c.GameMode.SetClient(client)
	c.World.SetClient(client)
	c.Status.SetClient(client)
}

// init 启动时间
var initTime = time.Now()
