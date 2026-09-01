// Package builder 提供建筑生成功能。
package builder

import (
	"fmt"
	"strings"
)

// Block 块
type Block struct {
	X, Y, Z int
	Type    string
}

// Builder 建筑生成器
type Builder struct {
	blocks []Block
}

// New 创建 Builder
func New() *Builder {
	return &Builder{}
}

// Build 生成建筑
func (b *Builder) Build(structureType string, args map[string]interface{}) (string, error) {
	b.blocks = nil
	size := 10
	if s, ok := args["size"].(int); ok {
		size = s
	}
	x, y, z := 0, 64, 0
	if v, ok := args["x"].(int); ok {
		x = v
	}
	if v, ok := args["y"].(int); ok {
		y = v
	}
	if v, ok := args["z"].(int); ok {
		z = v
	}

	switch structureType {
	case "house":
		b.buildHouse(x, y, z, size)
	case "tower":
		b.buildTower(x, y, z, size)
	case "circle":
		b.buildCircle(x, y, z, size)
	case "sphere":
		b.buildSphere(x, y, z, size)
	case "wall":
		b.buildWall(x, y, z, size)
	case "floor":
		b.buildFloor(x, y, z, size)
	case "rect":
		b.buildRect(x, y, z, size)
	default:
		return "", fmt.Errorf("unknown structure type: %s", structureType)
	}

	return b.summary()
}

func (b *Builder) add(x, y, z int, t string) {
	b.blocks = append(b.blocks, Block{X: x, Y: y, Z: z, Type: t})
}

func (b *Builder) buildHouse(x, y, z, size int) {
	// 地板
	for dx := 0; dx < size; dx++ {
		for dz := 0; dz < size; dz++ {
			b.add(x+dx, y, z+dz, "stone")
		}
	}
	// 墙
	for h := 1; h < size/2; h++ {
		for dx := 0; dx < size; dx++ {
			b.add(x+dx, y+h, z, "oak_log")
			b.add(x+dx, y+h, z+size-1, "oak_log")
		}
		for dz := 0; dz < size; dz++ {
			b.add(x, y+h, z+dz, "oak_log")
			b.add(x+size-1, y+h, z+dz, "oak_log")
		}
	}
	// 屋顶
	for h := 0; h < size/2; h++ {
		inset := h
		for dx := inset; dx < size-inset; dx++ {
			b.add(x+dx, y+size/2+h, z+inset, "oak_planks")
			b.add(x+dx, y+size/2+h, z+size-1-inset, "oak_planks")
		}
		for dz := inset; dz < size-inset; dz++ {
			b.add(x+inset, y+size/2+h, z+dz, "oak_planks")
			b.add(x+size-1-inset, y+size/2+h, z+dz, "oak_planks")
		}
	}
	// 门
	b.add(x+size/2, y+1, z, "air")
	b.add(x+size/2, y+2, z, "air")
	// 窗
	for h := 2; h < size/2-1; h++ {
		b.add(x+size/4, y+h, z, "glass")
		b.add(x+size*3/4, y+h, z, "glass")
	}
}

func (b *Builder) buildTower(x, y, z, size int) {
	height := size * 3
	for h := 0; h < height; h++ {
		for dx := 0; dx < size; dx++ {
			for dz := 0; dz < size; dz++ {
				isEdge := dx == 0 || dx == size-1 || dz == 0 || dz == size-1
				if isEdge {
					b.add(x+dx, y+h, z+dz, "stone_bricks")
				} else if h == 0 {
					b.add(x+dx, y+h, z+dz, "stone_bricks")
				}
			}
		}
	}
}

func (b *Builder) buildCircle(x, y, z, radius int) {
	for dx := -radius; dx <= radius; dx++ {
		for dz := -radius; dz <= radius; dz++ {
			distSq := dx*dx + dz*dz
			if distSq <= radius*radius {
				b.add(x+dx, y, z+dz, "stone")
			}
		}
	}
}

func (b *Builder) buildSphere(x, y, z, radius int) {
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			for dz := -radius; dz <= radius; dz++ {
				distSq := dx*dx + dy*dy + dz*dz
				if distSq <= radius*radius && distSq >= (radius-1)*(radius-1) {
					b.add(x+dx, y+dy+radius, z+dz, "glass")
				}
			}
		}
	}
}

func (b *Builder) buildWall(x, y, z, size int) {
	for h := 0; h < size; h++ {
		for dx := 0; dx < size; dx++ {
			b.add(x+dx, y+h, z, "cobblestone")
		}
	}
}

func (b *Builder) buildFloor(x, y, z, size int) {
	for dx := 0; dx < size; dx++ {
		for dz := 0; dz < size; dz++ {
			b.add(x+dx, y, z+dz, "oak_planks")
		}
	}
}

func (b *Builder) buildRect(x, y, z, size int) {
	for dx := 0; dx < size; dx++ {
		for dz := 0; dz < size; dz++ {
			for h := 0; h < size/2; h++ {
				if dx == 0 || dx == size-1 || dz == 0 || dz == size-1 || h == 0 {
					b.add(x+dx, y+h, z+dz, "bricks")
				}
			}
		}
	}
}

func (b *Builder) summary() (string, error) {
	if len(b.blocks) == 0 {
		return "", fmt.Errorf("no blocks generated")
	}

	// 统计每种块
	counts := make(map[string]int)
	for _, blk := range b.blocks {
		counts[blk.Type]++
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ 生成 %d 个块（%d 种）\n\n", len(b.blocks), len(counts)))
	for t, c := range counts {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", t, c))
	}
	return sb.String(), nil
}
