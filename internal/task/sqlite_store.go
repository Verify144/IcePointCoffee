package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SQLiteStore 基于 SQLite 的任务持久化存储。
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore 创建 SQLite 存储。
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	s := &SQLiteStore{db: db}
	s.migrate()
	return s
}

func (s *SQLiteStore) migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL DEFAULT '',
		payload TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL DEFAULT 'pending',
		priority INTEGER NOT NULL DEFAULT 2,
		progress INTEGER NOT NULL DEFAULT 0,
		result TEXT NOT NULL DEFAULT '',
		error TEXT NOT NULL DEFAULT '',
		retries INTEGER NOT NULL DEFAULT 0,
		max_retries INTEGER NOT NULL DEFAULT 3,
		created_at DATETIME NOT NULL,
		started_at DATETIME,
		completed_at DATETIME,
		user_id TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '[]'
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);
	`
	s.db.Exec(schema)
}

// Create 创建任务
func (s *SQLiteStore) Create(t *Task) error {
	payload, _ := json.Marshal(t.Payload)
	tags, _ := json.Marshal(t.Tags)

	var startedAt, completedAt *time.Time
	if t.StartedAt != nil {
		startedAt = t.StartedAt
	}
	if t.CompletedAt != nil {
		completedAt = t.CompletedAt
	}

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO tasks (id,type,payload,status,priority,progress,result,error,retries,max_retries,created_at,started_at,completed_at,user_id,tags)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.Type, string(payload), string(t.Status), t.Priority, t.Progress,
		t.Result, t.Error, t.Retries, t.MaxRetries,
		t.CreatedAt, startedAt, completedAt, t.UserID, string(tags),
	)
	return err
}

// Update 更新任务
func (s *SQLiteStore) Update(t *Task) error {
	payload, _ := json.Marshal(t.Payload)
	tags, _ := json.Marshal(t.Tags)

	var startedAt, completedAt *time.Time
	if t.StartedAt != nil {
		startedAt = t.StartedAt
	}
	if t.CompletedAt != nil {
		completedAt = t.CompletedAt
	}

	_, err := s.db.ExecContext(context.Background(),
		`UPDATE tasks SET type=?,payload=?,status=?,priority=?,progress=?,result=?,error=?,retries=?,max_retries=?,started_at=?,completed_at=?,user_id=?,tags=?
		 WHERE id=?`,
		t.Type, string(payload), string(t.Status), t.Priority, t.Progress,
		t.Result, t.Error, t.Retries, t.MaxRetries,
		startedAt, completedAt, t.UserID, string(tags), t.ID,
	)
	return err
}

// Get 根据 ID 获取任务
func (s *SQLiteStore) Get(id string) (*Task, error) {
	var t Task
	var payloadStr, tagsStr string
	var startedAt, completedAt sql.NullTime

	err := s.db.QueryRowContext(context.Background(),
		`SELECT id,type,payload,status,priority,progress,result,error,retries,max_retries,created_at,started_at,completed_at,user_id,tags
		 FROM tasks WHERE id=?`, id).Scan(
		&t.ID, &t.Type, &payloadStr, &t.Status, &t.Priority, &t.Progress,
		&t.Result, &t.Error, &t.Retries, &t.MaxRetries,
		&t.CreatedAt, &startedAt, &completedAt, &t.UserID, &tagsStr,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task %s not found", id)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(payloadStr), &t.Payload)
	if t.Payload == nil {
		t.Payload = make(map[string]interface{})
	}
	json.Unmarshal([]byte(tagsStr), &t.Tags)
	if startedAt.Valid {
		t.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}

	return &t, nil
}

// List 列出任务（支持过滤）
func (s *SQLiteStore) List(filter ListFilter) ([]*Task, error) {
	query := `SELECT id,type,payload,status,priority,progress,result,error,retries,max_retries,created_at,started_at,completed_at,user_id,tags FROM tasks WHERE 1=1`
	var args []interface{}

	if filter.Status != "" {
		query += " AND status=?"
		args = append(args, filter.Status)
	}
	if filter.Type != "" {
		query += " AND type=?"
		args = append(args, filter.Type)
	}
	if filter.UserID != "" {
		query += " AND user_id=?"
		args = append(args, filter.UserID)
	}

	query += " ORDER BY priority DESC, created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := s.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		var payloadStr, tagsStr string
		var startedAt, completedAt sql.NullTime

		if err := rows.Scan(
			&t.ID, &t.Type, &payloadStr, &t.Status, &t.Priority, &t.Progress,
			&t.Result, &t.Error, &t.Retries, &t.MaxRetries,
			&t.CreatedAt, &startedAt, &completedAt, &t.UserID, &tagsStr,
		); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(payloadStr), &t.Payload)
		if t.Payload == nil {
			t.Payload = make(map[string]interface{})
		}
		json.Unmarshal([]byte(tagsStr), &t.Tags)
		if startedAt.Valid {
			t.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, &t)
	}
	return tasks, nil
}

// Delete 删除任务
func (s *SQLiteStore) Delete(id string) error {
	_, err := s.db.ExecContext(context.Background(), "DELETE FROM tasks WHERE id=?", id)
	return err
}

// Count 返回任务总数
func (s *SQLiteStore) Count() (int, error) {
	var n int
	err := s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&n)
	return n, err
}

// Stats 返回任务统计
func (s *SQLiteStore) Stats() (map[string]int, error) {
	stats := make(map[string]int)
	rows, err := s.db.QueryContext(context.Background(), "SELECT status, COUNT(*) FROM tasks GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}

	// 总数
	var total int
	s.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM tasks").Scan(&total)
	stats["total"] = total

	return stats, nil
}

// ListAll 列出全部任务（便捷方法）
func (s *SQLiteStore) ListAll(limit int) ([]*Task, error) {
	return s.List(ListFilter{Limit: limit})
}

// 辅助：join 字符串切片
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
