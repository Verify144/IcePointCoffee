// Package builder 负责将建筑需求转换为可执行指令。
// 冰点咖啡内置的简易建筑生成器，复杂结构由插件处理。
package builder

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// BuildRequest 建筑请求。
type BuildRequest struct {
	Type        string  `json:"type"`        // house | tower | circle | sphere | wall | floor
	CenterX     int     `json:"center_x"`
	CenterY     int     `json:"center_y"`
	CenterZ     int     `json:"center_z"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Depth       int     `json:"depth"`
	Radius      int     `json:"radius"`
	BlockName   string  `json:"block_name"`
	Roof        bool    `json:"roof"`
	Hollow      bool    `json:"hollow"`
}

// BuildResponse 建筑结果。
type BuildResponse struct {
	Commands    []string `json:"commands"`
	BlockCount  int      `json:"block_count"`
	Description string   `json:"description"`
}

// Build 生成建筑指令。
func Build(req BuildRequest) (*BuildResponse, error) {
	if req.BlockName == "" {
		req.BlockName = "minecraft:stone"
	}
	if !strings.Contains(req.BlockName, ":") {
		req.BlockName = "minecraft:" + req.BlockName
	}

	switch req.Type {
	case "house":
		return buildHouse(req)
	case "tower":
		return buildTower(req)
	case "circle":
		return buildCircle(req)
	case "sphere":
		return buildSphere(req)
	case "wall":
		return buildWall(req)
	case "floor":
		return buildFloor(req)
	case "rect":
		return buildRect(req)
	default:
		return nil, fmt.Errorf("不支持的建筑类型: %s", req.Type)
	}
}

// buildHouse 小屋。
func buildHouse(req BuildRequest) (*BuildResponse, error) {
	if req.Width == 0 {
		req.Width = 5
	}
	if req.Depth == 0 {
		req.Depth = 5
	}
	if req.Height == 0 {
		req.Height = 4
	}
	x1 := req.CenterX
	x2 := req.CenterX + req.Width - 1
	z1 := req.CenterZ
	z2 := req.CenterZ + req.Depth - 1
	y1 := req.CenterY
	y2 := req.CenterY + req.Height - 1

	var cmds []string
	cmds = append(cmds, fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, y1, z1, x2, y1, z2, req.BlockName))
	cmds = append(cmds, fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, y1, z1, x1, y2, z2, req.BlockName))
	cmds = append(cmds, fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x2, y1, z1, x2, y2, z2, req.BlockName))
	cmds = append(cmds, fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, y1, z1, x2, y2, z1, req.BlockName))
	cmds = append(cmds, fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, y1, z2, x2, y2, z2, req.BlockName))

	count := req.Width*req.Depth + 2*req.Height*(req.Width+req.Depth)

	return &BuildResponse{
		Commands:    cmds,
		BlockCount:  count,
		Description: fmt.Sprintf("小屋 %dx%dx%d", req.Width, req.Height, req.Depth),
	}, nil
}

// buildTower 高塔。
func buildTower(req BuildRequest) (*BuildResponse, error) {
	if req.Width == 0 {
		req.Width = 3
	}
	if req.Depth == 0 {
		req.Depth = 3
	}
	if req.Height == 0 {
		req.Height = 20
	}
	x1 := req.CenterX
	x2 := req.CenterX + req.Width - 1
	z1 := req.CenterZ
	z2 := req.CenterZ + req.Depth - 1

	cmd := fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, req.CenterY, z1, x2, req.CenterY+req.Height-1, z2, req.BlockName)

	count := req.Width * req.Depth * req.Height

	return &BuildResponse{
		Commands:    []string{cmd},
		BlockCount:  count,
		Description: fmt.Sprintf("高塔 %dx%dx%d", req.Width, req.Height, req.Depth),
	}, nil
}

// buildCircle 圆形平台。
func buildCircle(req BuildRequest) (*BuildResponse, error) {
	if req.Radius == 0 {
		req.Radius = 5
	}
	r := float64(req.Radius)
	var cmds []string
	count := 0
	// 用 setblock 一个一个放
	for dx := -req.Radius; dx <= req.Radius; dx++ {
		for dz := -req.Radius; dz <= req.Radius; dz++ {
			d := math.Sqrt(float64(dx*dx) + float64(dz*dz))
			if d <= r+0.5 {
				x := req.CenterX + dx
				z := req.CenterZ + dz
				cmds = append(cmds, fmt.Sprintf("setblock %d %d %d %s",
					x, req.CenterY, z, req.BlockName))
				count++
			}
		}
	}
	return &BuildResponse{
		Commands:    cmds,
		BlockCount:  count,
		Description: fmt.Sprintf("圆形平台 半径=%d", req.Radius),
	}, nil
}

// buildSphere 球体。
func buildSphere(req BuildRequest) (*BuildResponse, error) {
	if req.Radius == 0 {
		req.Radius = 5
	}
	r := float64(req.Radius)
	var cmds []string
	count := 0
	for dx := -req.Radius; dx <= req.Radius; dx++ {
		for dy := -req.Radius; dy <= req.Radius; dy++ {
			for dz := -req.Radius; dz <= req.Radius; dz++ {
				d := math.Sqrt(float64(dx*dx) + float64(dy*dy) + float64(dz*dz))
				if d <= r+0.5 {
					x := req.CenterX + dx
					y := req.CenterY + dy
					z := req.CenterZ + dz
					cmds = append(cmds, fmt.Sprintf("setblock %d %d %d %s",
						x, y, z, req.BlockName))
					count++
				}
			}
		}
	}
	return &BuildResponse{
		Commands:    cmds,
		BlockCount:  count,
		Description: fmt.Sprintf("球体 半径=%d", req.Radius),
	}, nil
}

// buildWall 墙。
func buildWall(req BuildRequest) (*BuildResponse, error) {
	if req.Width == 0 {
		req.Width = 10
	}
	if req.Height == 0 {
		req.Height = 5
	}
	x1 := req.CenterX
	x2 := req.CenterX + req.Width - 1
	z := req.CenterZ
	cmd := fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, req.CenterY, z, x2, req.CenterY+req.Height-1, z, req.BlockName)

	count := req.Width * req.Height

	return &BuildResponse{
		Commands:    []string{cmd},
		BlockCount:  count,
		Description: fmt.Sprintf("墙 %dx%d", req.Width, req.Height),
	}, nil
}

// buildFloor 地板。
func buildFloor(req BuildRequest) (*BuildResponse, error) {
	if req.Width == 0 {
		req.Width = 10
	}
	if req.Depth == 0 {
		req.Depth = 10
	}
	x1 := req.CenterX
	x2 := req.CenterX + req.Width - 1
	z1 := req.CenterZ
	z2 := req.CenterZ + req.Depth - 1
	cmd := fmt.Sprintf("fill %d %d %d %d %d %d %s",
		x1, req.CenterY, z1, x2, req.CenterY, z2, req.BlockName)

	count := req.Width * req.Depth

	return &BuildResponse{
		Commands:    []string{cmd},
		BlockCount:  count,
		Description: fmt.Sprintf("地板 %dx%d", req.Width, req.Depth),
	}, nil
}

// buildRect 矩形区域。
func buildRect(req BuildRequest) (*BuildResponse, error) {
	if req.Width == 0 {
		req.Width = 10
	}
	if req.Height == 0 {
		req.Height = 5
	}
	if req.Depth == 0 {
		req.Depth = 10
	}
	x1 := req.CenterX
	x2 := req.CenterX + req.Width - 1
	y1 := req.CenterY
	y2 := req.CenterY + req.Height - 1
	z1 := req.CenterZ
	z2 := req.CenterZ + req.Depth - 1

	cmd := fmt.Sprintf("fill %d %d %d %d %d %d %s", x1, y1, z1, x2, y2, z2, req.BlockName)

	count := req.Width * req.Height * req.Depth

	return &BuildResponse{
		Commands:    []string{cmd},
		BlockCount:  count,
		Description: fmt.Sprintf("矩形 %dx%dx%d", req.Width, req.Height, req.Depth),
	}, nil
}

// ParseRequest 从指令字符串解析建筑请求。
// 格式: "type:house width:5 height:4 depth:5 block:oak_planks center:0,64,0"
func ParseRequest(s string) (BuildRequest, error) {
	req := BuildRequest{}
	parts := strings.Fields(s)
	for _, p := range parts {
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k, v := kv[0], kv[1]
		switch k {
		case "type":
			req.Type = v
		case "block":
			req.BlockName = v
		case "width", "w":
			n, _ := strconv.Atoi(v)
			req.Width = n
		case "height", "h":
			n, _ := strconv.Atoi(v)
			req.Height = n
		case "depth", "d":
			n, _ := strconv.Atoi(v)
			req.Depth = n
		case "radius", "r":
			n, _ := strconv.Atoi(v)
			req.Radius = n
		case "center":
			xyz := strings.Split(v, ",")
			if len(xyz) == 3 {
				req.CenterX, _ = strconv.Atoi(xyz[0])
				req.CenterY, _ = strconv.Atoi(xyz[1])
				req.CenterZ, _ = strconv.Atoi(xyz[2])
			}
		}
	}
	if req.Type == "" {
		return req, fmt.Errorf("未指定 type")
	}
	return req, nil
}
