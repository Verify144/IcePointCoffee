// Package db 提供 SQLite 持久化层。
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB 数据库连接。
type DB struct {
	sqldb *sql.DB
}

// New 打开 SQLite 数据库。
func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite 最佳实践

	// 建表
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	d := &DB{sqldb: db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("迁移失败: %w", err)
	}

	return d, nil
}

func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		prompt TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'command',
		description TEXT,
		result TEXT,
		commands TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		error TEXT,
		created_at DATETIME NOT NULL,
		done_at DATETIME
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

	CREATE TABLE IF NOT EXISTS configs (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS plugins (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1,
		config TEXT,
		loaded_at DATETIME
	);
	`
	_, err := d.sqldb.Exec(schema)
	return err
}

// Close 关闭数据库。
func (d *DB) Close() error {
	return d.sqldb.Close()
}

// TaskStore 任务存储。
type TaskStore struct {
	db *DB
}

// NewTaskStore 创建任务存储。
func NewTaskStore(db *DB) *TaskStore {
	return &TaskStore{db: db}
}

// Task 任务。
type Task struct {
	ID          string
	UserID      string
	Prompt      string
	Type        string  // command | structure | import
	Description string
	Result      string
	Commands    []string
	Status      string  // pending | running | done | failed
	Error       string
	CreatedAt   time.Time
	DoneAt      time.Time
}

// Create 创建任务。
func (s *TaskStore) Create(t *Task) error {
	commands := ""
	for _, c := range t.Commands {
		commands += c + "\n"
	}
	_, err := s.db.sqldb.ExecContext(context.Background(),
		`INSERT INTO tasks (id,user_id,prompt,type,description,result,commands,status,error,created_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.UserID, t.Prompt, t.Type, t.Description, t.Result, commands, t.Status, t.Error, t.CreatedAt,
	)
	return err
}

// Update 更新任务。
func (s *TaskStore) Update(t *Task) error {
	commands := ""
	for _, c := range t.Commands {
		commands += c + "\n"
	}
	_, err := s.db.sqldb.ExecContext(context.Background(),
		`UPDATE tasks SET type=?,description=?,result=?,commands=?,status=?,error=?,done_at=? WHERE id=?`,
		t.Type, t.Description, t.Result, commands, t.Status, t.Error, t.DoneAt, t.ID,
	)
	return err
}

// GetByID 根据 ID 获取任务。
func (s *TaskStore) GetByID(id string) (*Task, error) {
	var t Task
	var commands string
	err := s.db.sqldb.QueryRowContext(context.Background(),
		`SELECT id,user_id,prompt,type,description,result,commands,status,error,created_at,done_at
		 FROM tasks WHERE id=?`, id).Scan(
		&t.ID, &t.UserID, &t.Prompt, &t.Type, &t.Description, &t.Result, &commands,
		&t.Status, &t.Error, &t.CreatedAt, &t.DoneAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("任务不存在: %s", id)
	}
	if err != nil {
		return nil, err
	}
	for _, c := range splitLines(commands) {
		if c != "" {
			t.Commands = append(t.Commands, c)
		}
	}
	return &t, nil
}

// ListByUser 列出用户的任务。
func (s *TaskStore) ListByUser(userID string, limit int) ([]*Task, error) {
	rows, err := s.db.sqldb.QueryContext(context.Background(),
		`SELECT id,user_id,prompt,type,description,result,commands,status,error,created_at,done_at
		 FROM tasks WHERE user_id=? ORDER BY created_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		var commands string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Prompt, &t.Type, &t.Description, &t.Result, &commands,
			&t.Status, &t.Error, &t.CreatedAt, &t.DoneAt); err != nil {
			return nil, err
		}
		for _, c := range splitLines(commands) {
			if c != "" {
				t.Commands = append(t.Commands, c)
			}
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

// ListAll 列出所有任务。
func (s *TaskStore) ListAll(limit int) ([]*Task, error) {
	rows, err := s.db.sqldb.QueryContext(context.Background(),
		`SELECT id,user_id,prompt,type,description,result,commands,status,error,created_at,done_at
		 FROM tasks ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		var commands string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Prompt, &t.Type, &t.Description, &t.Result, &commands,
			&t.Status, &t.Error, &t.CreatedAt, &t.DoneAt); err != nil {
			return nil, err
		}
		for _, c := range splitLines(commands) {
			if c != "" {
				t.Commands = append(t.Commands, c)
			}
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	return lines
}
