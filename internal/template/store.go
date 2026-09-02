// Package template 提供建筑模板存储和管理。
package template

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// Template 建筑模板
type Template struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Category     string                 `json:"category"`
	ParamsSchema map[string]interface{} `json:"params_schema"` // 参数定义（类型/默认值/范围）
	Blocks       map[string]interface{} `json:"blocks"`        // 方块数据（可以是 NBT/结构JSON）
	IsPublic     bool                   `json:"is_public"`
	Likes        int                    `json:"likes"`
	Uses         int                    `json:"uses"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at,omitempty"`
}

// Store 模板存储
type Store struct {
	db *sql.DB
}

// NewStore 创建模板存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create 创建模板
func (s *Store) Create(t *Template) error {
	if t.ID == "" {
		t.ID = generateTemplateID()
	}
	paramsJSON, _ := json.Marshal(t.ParamsSchema)
	blocksJSON, _ := json.Marshal(t.Blocks)

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO build_templates (id, user_id, name, description, category, params_schema, blocks, is_public, likes, uses, created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.UserID, t.Name, t.Description, t.Category,
		string(paramsJSON), string(blocksJSON),
		boolToInt(t.IsPublic), t.Likes, t.Uses, t.CreatedAt)
	return err
}

// Get 获取模板
func (s *Store) Get(id string) (*Template, error) {
	var t Template
	var paramsJSON, blocksJSON string
	var updatedAt sql.NullTime

	err := s.db.QueryRowContext(context.Background(),
		`SELECT id, user_id, name, description, category, params_schema, blocks, is_public, likes, uses, created_at, updated_at
		 FROM build_templates WHERE id=?`, id).
		Scan(&t.ID, &t.UserID, &t.Name, &t.Description, &t.Category,
			&paramsJSON, &blocksJSON, &t.IsPublic, &t.Likes, &t.Uses, &t.CreatedAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("template not found")
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(paramsJSON), &t.ParamsSchema)
	json.Unmarshal([]byte(blocksJSON), &t.Blocks)
	if updatedAt.Valid {
		t.UpdatedAt = updatedAt.Time
	}

	return &t, nil
}

// ListFilter 列表过滤条件
type ListFilter struct {
	UserID    string
	Category  string
	Public    *bool // nil=全部，true=仅公开，false=仅私有
	Search    string // 名称/描述搜索
	SortBy    string // "likes", "uses", "created_at"
	Limit     int
	Offset    int
}

// List 列出模板
func (s *Store) List(filter ListFilter) ([]*Template, error) {
	query := `SELECT id, user_id, name, description, category, params_schema, blocks, is_public, likes, uses, created_at, updated_at
	          FROM build_templates WHERE 1=1`
	var args []interface{}

	if filter.UserID != "" {
		query += " AND user_id=?"
		args = append(args, filter.UserID)
	}
	if filter.Category != "" {
		query += " AND category=?"
		args = append(args, filter.Category)
	}
	if filter.Public != nil {
		query += " AND is_public=?"
		args = append(args, boolToInt(*filter.Public))
	}
	if filter.Search != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	sort := "created_at DESC"
	switch filter.SortBy {
	case "likes":
		sort = "likes DESC"
	case "uses":
		sort = "uses DESC"
	case "created_at":
		sort = "created_at DESC"
	}
	query += " ORDER BY " + sort

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	query += fmt.Sprintf(" LIMIT %d", limit)
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := s.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*Template
	for rows.Next() {
		var t Template
		var paramsJSON, blocksJSON string
		var updatedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Description, &t.Category,
			&paramsJSON, &blocksJSON, &t.IsPublic, &t.Likes, &t.Uses, &t.CreatedAt, &updatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(paramsJSON), &t.ParamsSchema)
		json.Unmarshal([]byte(blocksJSON), &t.Blocks)
		if updatedAt.Valid {
			t.UpdatedAt = updatedAt.Time
		}
		templates = append(templates, &t)
	}
	return templates, rows.Err()
}

// Update 更新模板
func (s *Store) Update(t *Template) error {
	paramsJSON, _ := json.Marshal(t.ParamsSchema)
	blocksJSON, _ := json.Marshal(t.Blocks)

	_, err := s.db.ExecContext(context.Background(),
		`UPDATE build_templates SET name=?, description=?, category=?, params_schema=?, blocks=?, is_public=?, likes=?, uses=?, updated_at=?
		 WHERE id=?`,
		t.Name, t.Description, t.Category, string(paramsJSON), string(blocksJSON),
		boolToInt(t.IsPublic), t.Likes, t.Uses, time.Now(), t.ID)
	return err
}

// Delete 删除模板
func (s *Store) Delete(id, userID string) error {
	res, err := s.db.ExecContext(context.Background(),
		`DELETE FROM build_templates WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("template not found or not owned")
	}
	return nil
}

// IncrementUses 增加使用次数
func (s *Store) IncrementUses(id string) error {
	_, err := s.db.ExecContext(context.Background(),
		`UPDATE build_templates SET uses=uses+1 WHERE id=?`, id)
	return err
}

// IncrementLikes 增加点赞数
func (s *Store) IncrementLikes(id string) error {
	_, err := s.db.ExecContext(context.Background(),
		`UPDATE build_templates SET likes=likes+1 WHERE id=?`, id)
	return err
}

// Categories 返回所有分类及数量
func (s *Store) Categories() ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT category, COUNT(*) as count FROM build_templates WHERE is_public=1 GROUP BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var cat string
		var count int
		rows.Scan(&cat, &count)
		result = append(result, map[string]interface{}{"category": cat, "count": count})
	}
	return result, rows.Err()
}

// SeedPublicTemplates 插入一些公开示例模板（幂等）
func (s *Store) SeedPublicTemplates() error {
	// 检查是否已有公开模板
	var n int
	s.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM build_templates WHERE is_public=1`).Scan(&n)
	if n > 0 {
		return nil
	}

	templates := getPublicTemplates()
	for _, t := range templates {
		s.Create(t)
	}
	return nil
}

// ==== 工具函数 ====

func generateTemplateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("tpl_%x", b)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// getPublicTemplates 返回预置公开模板
func getPublicTemplates() []*Template {
	now := time.Now()
	return []*Template{
		{
			ID:          "tpl_house_001",
			UserID:      "system",
			Name:        "简约房屋",
			Description: "基础的 10x10 单层房屋，带门和窗户",
			Category:    "residential",
			ParamsSchema: map[string]interface{}{
				"width":  map[string]interface{}{"type": "int", "default": 10, "min": 5, "max": 30},
				"height": map[string]interface{}{"type": "int", "default": 5, "min": 3, "max": 15},
				"depth":  map[string]interface{}{"type": "int", "default": 10, "min": 5, "max": 30},
				"block":  map[string]interface{}{"type": "string", "default": "cobblestone"},
			},
			Blocks: map[string]interface{}{
				"type": "fill_box",
				"x1":   0, "y1": 0, "z1": 0,
				"x2":   "width-1", "y2": "height-1", "z2": "depth-1",
				"block": "block",
				"hollow": true,
			},
			IsPublic:  true,
			CreatedAt: now,
		},
		{
			ID:          "tpl_tower_001",
			UserID:      "system",
			Name:        "圆形塔楼",
			Description: "逐层向上的圆柱形塔楼",
			Category:    "structural",
			ParamsSchema: map[string]interface{}{
				"height": map[string]interface{}{"type": "int", "default": 20, "min": 5, "max": 100},
				"radius": map[string]interface{}{"type": "int", "default": 5, "min": 2, "max": 20},
				"block":  map[string]interface{}{"type": "string", "default": "stone_bricks"},
			},
			Blocks: map[string]interface{}{
				"type": "cylinder",
				"radius": "radius",
				"height": "height",
				"block":  "block",
			},
			IsPublic:  true,
			CreatedAt: now,
		},
		{
			ID:          "tpl_fountain_001",
			UserID:      "system",
			Name:        "中央喷泉",
			Description: "带有水池和中心柱的喷泉",
			Category:    "decoration",
			ParamsSchema: map[string]interface{}{
				"radius": map[string]interface{}{"type": "int", "default": 5, "min": 3, "max": 15},
				"height": map[string]interface{}{"type": "int", "default": 6, "min": 3, "max": 20},
				"water":  map[string]interface{}{"type": "string", "default": "water"},
			},
			Blocks: map[string]interface{}{
				"type": "fountain",
				"radius": "radius",
				"height": "height",
			},
			IsPublic:  true,
			CreatedAt: now,
		},
		{
			ID:          "tpl_bridge_001",
			UserID:      "system",
			Name:        "石桥",
			Description: "跨越河流或峡谷的石桥",
			Category:    "infrastructure",
			ParamsSchema: map[string]interface{}{
				"length": map[string]interface{}{"type": "int", "default": 20, "min": 5, "max": 100},
				"width":  map[string]interface{}{"type": "int", "default": 5, "min": 3, "max": 20},
				"block":  map[string]interface{}{"type": "string", "default": "cobblestone"},
			},
			Blocks: map[string]interface{}{
				"type": "bridge",
				"length": "length",
				"width":  "width",
				"block":  "block",
			},
			IsPublic:  true,
			CreatedAt: now,
		},
		{
			ID:          "tpl_castle_001",
			UserID:      "system",
			Name:        "城堡",
			Description: "带围墙和塔楼的完整城堡",
			Category:    "residential",
			ParamsSchema: map[string]interface{}{
				"size":  map[string]interface{}{"type": "int", "default": 30, "min": 20, "max": 100},
				"block": map[string]interface{}{"type": "string", "default": "stone_bricks"},
			},
			Blocks: map[string]interface{}{
				"type": "castle",
				"size":  "size",
				"block": "block",
			},
			IsPublic:  true,
			CreatedAt: now,
		},
	}
}
