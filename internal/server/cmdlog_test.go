package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"

	cmdstore "github.com/Verify144/IcePointCoffee/internal/db"
)

// newCmdLogDB 创建一个带 command_history 表的 db
func newCmdLogDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, _ := sql.Open("sqlite", filepath.Join(dir, "cmdlog.db")+"?_journal=WAL")
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE build_templates (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, category TEXT NOT NULL DEFAULT 'custom', params_schema TEXT NOT NULL DEFAULT '{}', blocks TEXT NOT NULL DEFAULT '{}', is_public INTEGER NOT NULL DEFAULT 0, likes INTEGER NOT NULL DEFAULT 0, uses INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL, updated_at DATETIME);
	CREATE TABLE cron_jobs (id TEXT PRIMARY KEY, name TEXT NOT NULL, cron_expr TEXT NOT NULL, task_type TEXT NOT NULL, payload TEXT NOT NULL DEFAULT '{}', enabled INTEGER NOT NULL DEFAULT 1, last_run_at DATETIME, last_run_status TEXT, next_run_at DATETIME, created_at DATETIME NOT NULL);
	CREATE TABLE command_history (id TEXT PRIMARY KEY, tool TEXT NOT NULL, command TEXT NOT NULL, args TEXT NOT NULL DEFAULT '{}', output TEXT NOT NULL DEFAULT '', success INTEGER NOT NULL DEFAULT 1, dangerous INTEGER NOT NULL DEFAULT 0, duration_ms INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL);
	`
	db.Exec(schema)
	return db
}

func TestPhase11_CmdLogAPI(t *testing.T) {
	db := newCmdLogDB(t)
	defer db.Close()

	s := NewServerWithDB(8080, db)

	// 初始应为空
	req := httptest.NewRequest("GET", "/api/v1/cmdlog", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List: %d", rec.Code)
	}

	// 手动添加 2 条历史
	store := cmdstore.NewCommandHistoryStore(db)
	store.Add(&cmdstore.CommandHistory{
		Tool: "mc_command", Command: "time set day",
		Output: "Time set", Success: true, DurationMs: 50, CreatedAt: time.Now(),
	})
	store.Add(&cmdstore.CommandHistory{
		Tool: "mc_fill", Command: "block=stone",
		Output: "1000 blocks changed", Success: true,
		Dangerous: true, DurationMs: 100, CreatedAt: time.Now(),
	})

	// 查询
	req = httptest.NewRequest("GET", "/api/v1/cmdlog", nil)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List: %d", rec.Code)
	}
	var resp struct {
		Commands []cmdstore.CommandHistory `json:"commands"`
		Count    int                 `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Count != 2 {
		t.Errorf("Expected 2, got %d", resp.Count)
	}

	// 清空
	req = httptest.NewRequest("DELETE", "/api/v1/cmdlog", nil)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Clear: %d", rec.Code)
	}

	// 应为空
	req = httptest.NewRequest("GET", "/api/v1/cmdlog", nil)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Count != 0 {
		t.Errorf("After clear, expected 0, got %d", resp.Count)
	}
}

func TestPhase11_CmdLogLimit(t *testing.T) {
	db := newCmdLogDB(t)
	defer db.Close()

	store := cmdstore.NewCommandHistoryStore(db)
	for i := 0; i < 10; i++ {
		store.Add(&cmdstore.CommandHistory{
			Tool: "x", Command: "y", CreatedAt: time.Now(),
		})
	}

	s := NewServerWithDB(8080, db)
	req := httptest.NewRequest("GET", "/api/v1/cmdlog?limit=3", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List: %d", rec.Code)
	}
	var resp struct {
		Count int `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Count != 3 {
		t.Errorf("Expected 3 with limit, got %d", resp.Count)
	}
}

func TestPhase11_ObserverIntegration(t *testing.T) {
	db := newCmdLogDB(t)
	defer db.Close()

	s := NewServerWithDB(8080, db)

	// 验证 cmdHistory 已初始化
	if s.cmdHistory == nil {
		t.Fatal("cmdHistory should be set")
	}

	// 验证初始为空
	req := httptest.NewRequest("GET", "/api/v1/cmdlog", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List: %d", rec.Code)
	}
	var resp struct {
		Count int `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Count != 0 {
		t.Errorf("Expected 0 initially, got %d", resp.Count)
	}

	// 验证 API 方法不支持
	req = httptest.NewRequest("PUT", "/api/v1/cmdlog", bytes.NewBufferString("{}"))
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", rec.Code)
	}
}

func TestPhase11_CmdLogInvalidMethod(t *testing.T) {
	db := newCmdLogDB(t)
	defer db.Close()

	s := NewServerWithDB(8080, db)
	req := httptest.NewRequest("PUT", "/api/v1/cmdlog", bytes.NewBufferString("{}"))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", rec.Code)
	}
}
