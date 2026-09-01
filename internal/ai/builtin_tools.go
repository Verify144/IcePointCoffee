package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EchoTool 回显工具
type EchoTool struct{}

func (t *EchoTool) Name() string { return "echo" }
func (t *EchoTool) Description() string { return "回显输入的文本" }
func (t *EchoTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"text": map[string]interface{}{"type": "string", "description": "要回显的文本"},
		},
		"required": []string{"text"},
	}
}

func (t *EchoTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct{ Text string }
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	return params.Text, nil
}

// GetTimeTool 获取时间工具
type GetTimeTool struct{}

func (t *GetTimeTool) Name() string { return "get_time" }
func (t *GetTimeTool) Description() string { return "获取当前时间" }
func (t *GetTimeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}

func (t *GetTimeTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	return time.Now().Format("2006-01-02 15:04:05 MST"), nil
}

// CalculateTool 计算工具
type CalculateTool struct{}

func (t *CalculateTool) Name() string { return "calculate" }
func (t *CalculateTool) Description() string { return "基础数学计算（加减乘除）" }
func (t *CalculateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{"type": "string", "description": "数学表达式"},
		},
		"required": []string{"expression"},
	}
}

func (t *CalculateTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct{ Expression string }
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	return simpleCalc(params.Expression)
}

// simpleCalc 简单表达式计算（仅支持 + - * / 和数字）
func simpleCalc(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	if len(expr) == 0 {
		return 0, fmt.Errorf("empty expression")
	}
	// 简单解析：递归 + -
	return parseAddSub(expr)
}

func parseAddSub(s string) (float64, error) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '+' {
			a, err := parseMulDiv(s[:i])
			if err != nil {
				return 0, err
			}
			b, err := parseAddSub(s[i+1:])
			if err != nil {
				return 0, err
			}
			return a + b, nil
		}
		if s[i] == '-' && i > 0 {
			a, err := parseMulDiv(s[:i])
			if err != nil {
				return 0, err
			}
			b, err := parseAddSub(s[i+1:])
			if err != nil {
				return 0, err
			}
			return a - b, nil
		}
	}
	return parseMulDiv(s)
}

func parseMulDiv(s string) (float64, error) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '*' {
			a, err := parseNum(s[:i])
			if err != nil {
				return 0, err
			}
			b, err := parseMulDiv(s[i+1:])
			if err != nil {
				return 0, err
			}
			return a * b, nil
		}
		if s[i] == '/' {
			a, err := parseNum(s[:i])
			if err != nil {
				return 0, err
			}
			b, err := parseMulDiv(s[i+1:])
			if err != nil {
				return 0, err
			}
			if b == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return a / b, nil
		}
	}
	return parseNum(s)
}

func parseNum(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// SendCommandTool 发送命令工具
type SendCommandTool struct {
	Executor interface {
		Execute(cmd string) (string, error)
	}
}

func (t *SendCommandTool) Name() string { return "send_command" }
func (t *SendCommandTool) Description() string { return "向 MC 服务器发送命令" }
func (t *SendCommandTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{"type": "string", "description": "MC 命令（不带 /）"},
		},
		"required": []string{"command"},
	}
}

func (t *SendCommandTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	if t.Executor == nil {
		return nil, fmt.Errorf("no executor configured")
	}
	var params struct{ Command string }
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	return t.Executor.Execute(params.Command)
}

// BuildStructureTool 生成建筑工具
type BuildStructureTool struct {
	Builder interface {
		Build(structureType string, args map[string]interface{}) (string, error)
	}
}

func (t *BuildStructureTool) Name() string { return "build_structure" }
func (t *BuildStructureTool) Description() string { return "生成建筑（house/tower/circle/sphere/wall/floor/rect）" }
func (t *BuildStructureTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{"type": "string", "description": "建筑类型"},
			"size": map[string]interface{}{"type": "integer", "description": "尺寸"},
			"x":    map[string]interface{}{"type": "integer", "description": "X 坐标"},
			"y":    map[string]interface{}{"type": "integer", "description": "Y 坐标"},
			"z":    map[string]interface{}{"type": "integer", "description": "Z 坐标"},
		},
		"required": []string{"type"},
	}
}

func (t *BuildStructureTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	if t.Builder == nil {
		return nil, fmt.Errorf("no builder configured")
	}
	var params struct {
		Type string                 `json:"type"`
		Args map[string]interface{} `json:"args"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Args == nil {
		params.Args = make(map[string]interface{})
	}
	params.Args["type"] = params.Type
	return t.Builder.Build(params.Type, params.Args)
}
