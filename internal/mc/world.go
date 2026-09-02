package mc

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// WorldState 世界感知状态（AI 视角）
type WorldState struct {
	mu sync.RWMutex
	Player
}

// Player 当前玩家信息
type Player struct {
	Name      string  `json:"name"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	Health    float32 `json:"health"`
	MaxHealth float32 `json:"max_health"`
	Food      int     `json:"food"`
	Saturation float32 `json:"saturation"`
	GameMode  string  `json:"game_mode"`
	World     string  `json:"world"`
}

// NearbyPlayer 附近玩家
type NearbyPlayer struct {
	Name string `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	Dist float64 `json:"dist"`
}

// NearbyEntity 附近实体
type NearbyEntity struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	Dist float64 `json:"dist"`
}

// NearbyBlock 附近方块
type NearbyBlock struct {
	ID    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Z     int    `json:"z"`
	Dist  float64 `json:"dist"`
}

// PerceiveResult 感知结果
type PerceiveResult struct {
	Success bool `json:"success"`
	Player         Player          `json:"player"`
	NearbyPlayers  []NearbyPlayer  `json:"nearby_players"`
	NearbyEntities []NearbyEntity  `json:"nearby_entities"`
	NearbyBlocks   []NearbyBlock   `json:"nearby_blocks"`
	Description    string          `json:"description"`
}

// NewWorldState 创建感知引擎
func NewWorldState() *WorldState {
	return &WorldState{}
}

// UpdatePlayer 更新玩家信息
func (ws *WorldState) UpdatePlayer(p Player) {
	ws.mu.Lock()
	ws.Player = p
	ws.mu.Unlock()
}

// GetPlayer 获取当前玩家
func (ws *WorldState) GetPlayer() Player {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.Player
}

// Perceive 执行一次环境感知（执行多个命令收集数据）
func (ws *WorldState) Perceive(ctx context.Context, client ClientInterface, radius int, detail string) *PerceiveResult {
	if client == nil {
		return &PerceiveResult{Success: false, Description: "MC 客户端未连接"}
	}

	result := &PerceiveResult{
		Success:         true,
		NearbyPlayers:  []NearbyPlayer{},
		NearbyEntities: []NearbyEntity{},
		NearbyBlocks:   []NearbyBlock{},
	}

	// 1. 获取玩家状态
	if p := ws.queryPlayerState(ctx, client); p.Name != "" {
		result.Player = p
		ws.UpdatePlayer(p)
	}

	// 2. 获取附近玩家
	result.NearbyPlayers = ws.queryNearbyPlayers(ctx, client, radius)

	// 3. 获取附近实体（根据 detail 级别）
	if detail != "low" {
		result.NearbyEntities = ws.queryNearbyEntities(ctx, client, radius)
	}

	// 4. 简单描述生成
	result.Description = ws.generateDescription(result, radius)

	return result
}

// ---- 内部查询方法 ----

// queryPlayerState 查询玩家状态
func (ws *WorldState) queryPlayerState(ctx context.Context, client ClientInterface) Player {
	p := Player{Health: 20, MaxHealth: 20, Food: 20}

	// 位置
	out, _ := client.SendCommand(ctx, "data get entity @s Pos")
	if out != "" {
		parts := ws.parseVec3(out)
		if len(parts) >= 3 {
			p.X, _ = strconv.ParseFloat(parts[0], 64)
			p.Y, _ = strconv.ParseFloat(parts[1], 64)
			p.Z, _ = strconv.ParseFloat(parts[2], 64)
		}
	}

	// 生命值
	out, _ = client.SendCommand(ctx, "data get entity @s Health")
	if out != "" {
		if v := ws.parseFloat(out); v > 0 {
			p.Health = v
		}
	}

	// 饱食度
	out, _ = client.SendCommand(ctx, "data get entity @s foodLevel")
	if out != "" {
		if v := ws.parseInt(out); v > 0 {
			p.Food = v
		}
	}

	// 游戏模式
	out, _ = client.SendCommand(ctx, "gamemode query")
	if out != "" {
		if strings.Contains(out, "survival") {
			p.GameMode = "survival"
		} else if strings.Contains(out, "creative") {
			p.GameMode = "creative"
		} else if strings.Contains(out, "adventure") {
			p.GameMode = "adventure"
		} else if strings.Contains(out, "spectator") {
			p.GameMode = "spectator"
		} else {
			p.GameMode = strings.TrimSpace(out)
		}
	}

	return p
}

// queryNearbyPlayers 查询附近玩家
func (ws *WorldState) queryNearbyPlayers(ctx context.Context, client ClientInterface, radius int) []NearbyPlayer {
	var players []NearbyPlayer

	out, _ := client.SendCommand(ctx, "list")
	if out == "" {
		return players
	}

	// 解析 "There are X players online: Player1, Player2"
	re := regexp.MustCompile(`online:\s*(.*?)(?:\n|$)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) < 2 {
		return players
	}

	names := strings.Split(matches[1], ",")
	me := ws.GetPlayer()

	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// 查每个玩家的位置
		cmdOut, _ := client.SendCommand(ctx, fmt.Sprintf("data get entity %s Pos", name))
		parts := ws.parseVec3(cmdOut)
		if len(parts) >= 3 {
			x, _ := strconv.ParseFloat(parts[0], 64)
			y, _ := strconv.ParseFloat(parts[1], 64)
			z, _ := strconv.ParseFloat(parts[2], 64)
			dist := ws.dist(me.X, me.Y, me.Z, x, y, z)
			if dist <= float64(radius)+5 { // 放宽一点
				players = append(players, NearbyPlayer{
					Name: name,
					X:    x,
					Y:    y,
					Z:    z,
					Dist: dist,
				})
			}
		}
	}

	return players
}

// queryNearbyEntities 查询附近实体
func (ws *WorldState) queryNearbyEntities(ctx context.Context, client ClientInterface, radius int) []NearbyEntity {
	var entities []NearbyEntity

	// 使用 execute 检测实体
	out, _ := client.SendCommand(ctx, fmt.Sprintf("execute as @s at @s run testfor @e[distance=..%d]", radius))
	if out == "" {
		return entities
	}

	// 常见实体类型
	entityTypes := []string{"zombie", "skeleton", "creeper", "spider", "pig", "cow", "sheep", "chicken",
		"horse", "donkey", "mule", "villager", "wandering_trader", "iron_golem", "snow_golem",
		"shulker", "enderman", "blaze", "zombified_piglin", "ghast", "magma_cube",
		"vex", "evoker", "vindicator", "pillager", "ravager",
		"drowned", "husk", "stray", "phantom", "dolphin", "turtle",
		"cod", "salmon", "tropical_fish", "pufferfish", "axolotl",
		"glow_squid", "strider", "hoglin", "zoglin", "piglin", "piglin_brute",
		"wither_skeleton", "strider", "goat"}

	me := ws.GetPlayer()
	for _, etype := range entityTypes {
		// 查附近这类实体数量
		cmdOut, _ := client.SendCommand(ctx, fmt.Sprintf("execute as @s at @s if entity @e[type=%s,distance=..%d] run say __found__", etype, radius))
		if strings.Contains(cmdOut, "__found__") {
			// 查一个具体位置
			posOut, _ := client.SendCommand(ctx, fmt.Sprintf("execute as @s at @s as @e[type=%s,limit=1,distance=..%d] run data get entity @s Pos",
				etype, radius))
			parts := ws.parseVec3(posOut)
			var x, y, z, dist float64
			if len(parts) >= 3 {
				x, _ = strconv.ParseFloat(parts[0], 64)
				y, _ = strconv.ParseFloat(parts[1], 64)
				z, _ = strconv.ParseFloat(parts[2], 64)
				dist = ws.dist(me.X, me.Y, me.Z, x, y, z)
			}
			entities = append(entities, NearbyEntity{
				Type: etype,
				X:    x, Y: y, Z: z,
				Dist: dist,
			})
		}
	}

	return entities
}

// queryNearbyBlocks 查询附近方块（简化为特定坐标扫描）
func (ws *WorldState) queryNearbyBlocks(ctx context.Context, client ClientInterface, radius int) []NearbyBlock {
	var blocks []NearbyBlock
	me := ws.GetPlayer()

	// 用 testforblock 扫描附近地面
	// 简化：检测玩家脚下 3x3 区域
	y := int(me.Y) - 1
	for dx := -1; dx <= 1; dx++ {
		for dz := -1; dz <= 1; dz++ {
			x := int(me.X) + dx
			z := int(me.Z) + dz
			out, _ := client.SendCommand(ctx, fmt.Sprintf("testforblock %d %d %d *", x, y, z))
			if strings.Contains(out, "Successfully") || strings.Contains(out, "found") {
				dist := ws.dist(me.X, me.Y, me.Z, float64(x), float64(y), float64(z))
				if dist <= float64(radius) {
					// 尝试识别方块类型
					idOut, _ := client.SendCommand(ctx, fmt.Sprintf("testforblock %d %d %d", x, y, z))
					id := ws.identifyBlock(idOut)
					blocks = append(blocks, NearbyBlock{
						ID:   id,
						X:    x, Y: y, Z: z,
						Dist: dist,
					})
				}
			}
		}
	}

	return blocks
}

// generateDescription 生成自然语言描述
func (ws *WorldState) generateDescription(result *PerceiveResult, radius int) string {
	p := result.Player
	var b strings.Builder

	b.WriteString(fmt.Sprintf("你在世界「%s」中，坐标 (%.1f, %.1f, %.1f)，",
		p.World, p.X, p.Y, p.Z))

	b.WriteString(fmt.Sprintf("生命值 %.0f/%d，饱食度 %d/20，游戏模式：%s。",
		p.Health, int(p.MaxHealth), p.Food, p.GameMode))

	if len(result.NearbyPlayers) > 0 {
		b.WriteString(fmt.Sprintf("\n附近有 %d 名玩家：", len(result.NearbyPlayers)))
		for _, np := range result.NearbyPlayers {
			b.WriteString(fmt.Sprintf(" %s(距离%.0fm)", np.Name, np.Dist))
		}
		b.WriteString("。")
	}

	if len(result.NearbyEntities) > 0 {
		b.WriteString(fmt.Sprintf("\n附近有 %d 种生物：", len(result.NearbyEntities)))
		typeCount := make(map[string]int)
		for _, e := range result.NearbyEntities {
			typeCount[e.Type]++
		}
		for etype, cnt := range typeCount {
			b.WriteString(fmt.Sprintf(" %s(x%d)", etype, cnt))
		}
		b.WriteString("。")
	}

	if len(result.NearbyBlocks) > 0 {
		b.WriteString("\n脚下地面有方块：")
		blockSet := make(map[string]bool)
		for _, bl := range result.NearbyBlocks {
			blockSet[bl.ID] = true
		}
		for bid := range blockSet {
			b.WriteString(" " + bid)
		}
		b.WriteString("。")
	}

	if len(result.NearbyPlayers) == 0 && len(result.NearbyEntities) == 0 {
		b.WriteString("\n周围看起来很安静，没有其他玩家或生物。")
	}

	return b.String()
}

// ---- 辅助解析 ----

var vec3Re = regexp.MustCompile(`\[(-?[\d.]+),\s*(-?[\d.]+),\s*(-?[\d.]+)\]`)
var floatRe = regexp.MustCompile(`(-?[\d.]+)\s*(?:f|d)?$`)

func (ws *WorldState) parseVec3(s string) []string {
	m := vec3Re.FindStringSubmatch(s)
	if len(m) >= 4 {
		return m[1:]
	}
	return nil
}

func (ws *WorldState) parseFloat(s string) float32 {
	m := floatRe.FindStringSubmatch(s)
	if len(m) >= 2 {
		v, _ := strconv.ParseFloat(m[1], 32)
		return float32(v)
	}
	return 0
}

func (ws *WorldState) parseInt(s string) int {
	re := regexp.MustCompile(`(-?\d+)`)
	m := re.FindStringSubmatch(s)
	if len(m) >= 2 {
		v, _ := strconv.Atoi(m[1])
		return v
	}
	return 0
}

func (ws *WorldState) dist(x1, y1, z1, x2, y2, z2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	return (dx*dx + dy*dy + dz*dz)
}

func (ws *WorldState) identifyBlock(s string) string {
	if strings.Contains(s, "grass") {
		return "grass_block"
	}
	if strings.Contains(s, "stone") {
		return "stone"
	}
	if strings.Contains(s, "dirt") {
		return "dirt"
	}
	if strings.Contains(s, "water") {
		return "water"
	}
	if strings.Contains(s, "air") {
		return "air"
	}
	if strings.Contains(s, "sand") {
		return "sand"
	}
	return "unknown"
}

// MarshalJSON 序列化感知结果
func (r *PerceiveResult) MarshalJSON() ([]byte, error) {
	type Alias PerceiveResult
	return json.Marshal((*Alias)(r))
}
