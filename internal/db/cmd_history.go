package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"time"
)

// CommandHistory 单条命令历史
type CommandHistory struct {
	ID         string    `json:"id"`
	Tool       string    `json:"tool"`
	Command    string    `json:"command"`
	Args       string    `json:"args"`        // JSON string
	Output     string    `json:"output"`
	Success    bool      `json:"success"`
	Dangerous  bool      `json:"dangerous"`
	DurationMs int64     `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// CommandHistoryStore 命令历史存储
type CommandHistoryStore struct {
	db *sql.DB
}

// NewCommandHistoryStore 创建命令历史存储
func NewCommandHistoryStore(db *sql.DB) *CommandHistoryStore {
	return &CommandHistoryStore{db: db}
}

// Add 记录一条命令
func (s *CommandHistoryStore) Add(h *CommandHistory) error {
	if h.ID == "" {
		h.ID = generateID("cmd_", 8)
	}
	argsJSON, _ := json.Marshal(h.Args)
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO command_history (id, tool, command, args, output, success, dangerous, duration_ms, created_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		h.ID, h.Tool, h.Command, string(argsJSON), h.Output,
		boolToInt(h.Success), boolToInt(h.Dangerous), h.DurationMs, h.CreatedAt)
	return err
}

// List 列出最近 N 条
func (s *CommandHistoryStore) List(limit int) ([]*CommandHistory, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT id, tool, command, args, output, success, dangerous, duration_ms, created_at
		 FROM command_history ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*CommandHistory
	for rows.Next() {
		var h CommandHistory
		var argsJSON string
		if err := rows.Scan(&h.ID, &h.Tool, &h.Command, &argsJSON, &h.Output,
			&h.Success, &h.Dangerous, &h.DurationMs, &h.CreatedAt); err != nil {
			return nil, err
		}
		h.Args = argsJSON
		list = append(list, &h)
	}
	return list, rows.Err()
}

// Clear 清空历史
func (s *CommandHistoryStore) Clear() error {
	_, err := s.db.ExecContext(context.Background(), `DELETE FROM command_history`)
	return err
}

// Count 返回总数
func (s *CommandHistoryStore) Count() (int, error) {
	var n int
	err := s.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM command_history`).Scan(&n)
	return n, err
}

func generateID(prefix string, n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return prefix + hexEncode(b)
}

func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	r := make([]byte, len(b)*2)
	for i, v := range b {
		r[i*2] = hex[v>>4]
		r[i*2+1] = hex[v&0xf]
	}
	return string(r)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
