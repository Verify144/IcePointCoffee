package ai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Verify144/IcePointCoffee/internal/mc"
)

func TestIsBlacklisted(t *testing.T) {
	dangerous := []string{
		"stop", "kick Steve", "ban Steve", "op Steve",
		"deop Steve", "banlist", "whitelist off",
		"reload", "publish", "shutdown",
	}
	for _, cmd := range dangerous {
		if reason := IsBlacklisted(cmd); reason == "" {
			t.Errorf("'%s' should be blocked", cmd)
		}
	}

	safe := []string{
		"list", "time set day", "weather clear",
		"gamemode creative", "fill 0 64 0 10 70 10 stone",
		"setblock 0 64 0 diamond_block",
		"give Steve diamond 64",
	}
	for _, cmd := range safe {
		if reason := IsBlacklisted(cmd); reason != "" {
			t.Errorf("'%s' should be allowed, got: %s", cmd, reason)
		}
	}
}

func TestMCCommandToolSuccess(t *testing.T) {
	mcClient := mc.NewMock(true)
	cmd := NewMCCommandTool()
	cmd.SetClient(mcClient)

	args, _ := json.Marshal(map[string]string{"command": "list"})
	result, err := cmd.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("Should succeed")
	}
	if m["command"] != "list" {
		t.Errorf("Wrong command: %v", m["command"])
	}
}

func TestMCCommandToolBlacklisted(t *testing.T) {
	mcClient := mc.NewMock(true)
	cmd := NewMCCommandTool()
	cmd.SetClient(mcClient)

	args, _ := json.Marshal(map[string]string{"command": "stop"})
	result, err := cmd.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != false {
		t.Error("stop should be blocked")
	}
}

func TestMCFillToolVolumeLimit(t *testing.T) {
	mcClient := mc.NewMock(true)
	fill := NewMCFillTool()
	fill.SetClient(mcClient)

	args, _ := json.Marshal(map[string]interface{}{
		"x1": 0, "y1": 0, "z1": 0,
		"x2": 1000, "y2": 100, "z2": 100,
		"block": "stone",
	})
	result, err := fill.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != false {
		t.Error("Large fill should be blocked")
	}
}

func TestMCStatusTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	status := NewMCStatusTool()
	status.SetClient(mcClient)

	args, _ := json.Marshal(map[string]interface{}{})
	result, err := status.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	s := mc.Status(result.(mc.Status))
	if !s.Connected {
		t.Error("Mock should show connected")
	}
}

func TestMCChatTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	chat := NewMCChatTool()
	chat.SetClient(mcClient)

	args, _ := json.Marshal(map[string]string{"message": "Hello!"})
	result, err := chat.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("Chat should succeed")
	}
}

func TestMCTeleportTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	tp := NewMCTeleportTool()
	tp.SetClient(mcClient)

	args, _ := json.Marshal(map[string]interface{}{
		"target": "@s", "x": 100.0, "y": 64.0, "z": 200.0,
	})
	result, err := tp.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("Teleport should succeed")
	}
}

func TestMCGiveTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	give := NewMCGiveTool()
	give.SetClient(mcClient)

	args, _ := json.Marshal(map[string]interface{}{
		"target": "Steve", "item": "diamond", "count": 64,
	})
	result, err := give.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("Give should succeed")
	}
}

func TestMCSetBlockTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	sb := NewMCSetBlockTool()
	sb.SetClient(mcClient)

	args, _ := json.Marshal(map[string]interface{}{
		"x": 0, "y": 64, "z": 0, "block": "diamond_block",
	})
	result, err := sb.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("SetBlock should succeed")
	}
}

func TestMCWorldTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	world := NewMCWorldTool()
	world.SetClient(mcClient)

	// time
	args, _ := json.Marshal(map[string]string{"kind": "time", "value": "day"})
	result, err := world.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("time should succeed")
	}

	// weather
	args, _ = json.Marshal(map[string]string{"kind": "weather", "value": "clear"})
	result, err = world.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m = result.(map[string]interface{})
	if m["success"] != true {
		t.Error("weather should succeed")
	}

	// invalid
	args, _ = json.Marshal(map[string]string{"kind": "weather", "value": "snow"})
	result, _ = world.Execute(context.Background(), args)
	m = result.(map[string]interface{})
	if m["success"] != false {
		t.Error("invalid weather should fail")
	}
}

func TestMCDialogTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	dialog := NewMCDialogTool()
	dialog.SetClient(mcClient)

	args, _ := json.Marshal(map[string]interface{}{
		"target": "@a", "kind": "tellraw", "message": `"hi"`,
	})
	result, err := dialog.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("tellraw should succeed")
	}
}

func TestMCGameModeTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	gm := NewMCGameModeTool()
	gm.SetClient(mcClient)

	args, _ := json.Marshal(map[string]string{"target": "Steve", "mode": "creative"})
	result, err := gm.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("gamemode should succeed")
	}
}

func TestMCToolsNotConnected(t *testing.T) {
	mcClient := mc.NewMock(false)
	cmd := NewMCCommandTool()
	cmd.SetClient(mcClient)

	args, _ := json.Marshal(map[string]string{"command": "list"})
	_, err := cmd.Execute(context.Background(), args)
	if err == nil {
		t.Error("Should error when not connected")
	}
}

func TestMCControllerInject(t *testing.T) {
	mcClient := mc.NewMock(true)
	// 用零值 reg 测试 RegisterMCTools
	reg := NewToolRegistry()
	ctrl := RegisterMCTools(reg)
	ctrl.Inject(mcClient)

	// 所有工具已注册
	tools := reg.List()
	if len(tools) < 28 {
		t.Errorf("Expected at least 27 tools, got %d", len(tools))
	}

	// 找到 mc_command 并执行
	_, ok := reg.Get("mc_command")
	if !ok {
		t.Error("mc_command should be registered")
	}
}

func TestMCAttackTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	attack := NewMCAttackTool()
	attack.SetClient(mcClient)
	args, _ := json.Marshal(map[string]uint64{"target_id": 123})
	result, err := attack.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("attack should succeed")
	}
}

func TestMCPlaySoundTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	sound := NewMCPlaySoundTool()
	sound.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"sound_id": "random.levelup",
		"x": 0.0, "y": 64.0, "z": 0.0,
		"volume": 1.0, "pitch": 1.0,
	})
	result, err := sound.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("play_sound should succeed")
	}
}

func TestMCMakeParticlesTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	particle := NewMCMakeParticlesTool()
	particle.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"particle_id": 12, "x": 0.0, "y": 64.0, "z": 0.0,
	})
	result, err := particle.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("particle should succeed")
	}
}

func TestMCSpawnEntityTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	spawn := NewMCSpawnEntityTool()
	spawn.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"entity_type": "zombie", "x": 0.0, "y": 64.0, "z": 0.0,
	})
	result, err := spawn.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("spawn_entity should succeed")
	}
}

func TestMCSwingArmTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	swing := NewMCSwingArmTool()
	swing.SetClient(mcClient)
	result, err := swing.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("swing should succeed")
	}
}

func TestMCBossBarTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	bossbar := NewMCBossBarTool()
	bossbar.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"target": "@s", "title": "Boss!", "health": 1.0,
	})
	result, err := bossbar.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("bossbar should succeed")
	}
}

func TestMCRespawnTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	respawn := NewMCRespawnTool()
	respawn.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"x": 0.0, "y": 64.0, "z": 0.0,
	})
	result, err := respawn.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("respawn should succeed")
	}
}

func TestMCWeatherTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	w := NewMCWeatherTool()
	w.SetClient(mcClient)
	args, _ := json.Marshal(map[string]string{"weather": "rain"})
	result, err := w.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("weather should succeed")
	}
}

func TestMCWorldBorderTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	wb := NewMCWorldBorderTool()
	wb.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{"center_x": 0.0, "center_z": 0.0, "radius": 1000000.0})
	result, err := wb.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("worldborder should succeed")
	}
}

func TestMCSpawnpointTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	sp := NewMCSpawnpointTool()
	sp.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{"x": 0.0, "y": 64.0, "z": 0.0})
	result, err := sp.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("spawnpoint should succeed")
	}
}

func TestMCDifficultyTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	d := NewMCDifficultyTool()
	d.SetClient(mcClient)
	args, _ := json.Marshal(map[string]string{"difficulty": "hard"})
	result, err := d.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("difficulty should succeed")
	}
}

func TestMCNametagTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	n := NewMCNametagTool()
	n.SetClient(mcClient)
	args, _ := json.Marshal(map[string]string{"target": "zombie", "nametag": "Boss Zombie"})
	result, err := n.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("nametag should succeed")
	}
}

func TestMCXPTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	xp := NewMCXPTool()
	xp.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{"target": "@s", "amount": 100})
	result, err := xp.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("xp should succeed")
	}
}

func TestMCEffectTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	e := NewMCEffectTool()
	e.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{"target": "@s", "effect": "speed", "seconds": 60})
	result, err := e.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("effect should succeed")
	}
}

func TestMCClearInventoryTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	ci := NewMCClearInventoryTool()
	ci.SetClient(mcClient)
	result, err := ci.Execute(context.Background(), json.RawMessage(`{"target":"@s"}`))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != true {
		t.Error("clear_inventory should succeed")
	}
}

func TestMCPerceiveTool(t *testing.T) {
	mcClient := mc.NewMock(true)
	p := NewMCPerceiveTool()
	p.SetClient(mcClient)
	result, err := p.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(*mc.PerceiveResult)
	if !m.Success {
		t.Error("perceive should succeed")
	}
	if m.Description == "" {
		t.Error("description should not be empty")
	}
	if m.Player.Name == "" {
		t.Error("player name should not be empty")
	}
}

func TestMCPerceiveToolWithRadius(t *testing.T) {
	mcClient := mc.NewMock(true)
	p := NewMCPerceiveTool()
	p.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"radius": 20, "detail": "high", "scope": "nearby",
	})
	result, err := p.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(*mc.PerceiveResult)
	if !m.Success {
		t.Error("perceive with params should succeed")
	}
}

func TestMCPerceiveToolScopePlayers(t *testing.T) {
	mcClient := mc.NewMock(true)
	p := NewMCPerceiveTool()
	p.SetClient(mcClient)
	args, _ := json.Marshal(map[string]interface{}{
		"scope": "players",
	})
	result, err := p.Execute(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	m := result.(*mc.PerceiveResult)
	if !m.Success {
		t.Error("perceive players scope should succeed")
	}
}

func TestMCPerceiveToolNotConnected(t *testing.T) {
	mcClient := mc.NewMock(false)
	p := NewMCPerceiveTool()
	p.SetClient(mcClient)
	result, err := p.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	m := result.(map[string]interface{})
	if m["success"] != false {
		t.Error("perceive with no client should return error result")
	}
}
