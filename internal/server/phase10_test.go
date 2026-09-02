package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"database/sql"
	_ "modernc.org/sqlite"
)

func newServerDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, _ := sql.Open("sqlite", filepath.Join(dir, "phase10.db")+"?_journal=WAL")
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE build_templates (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT, category TEXT NOT NULL DEFAULT 'custom', params_schema TEXT NOT NULL DEFAULT '{}', blocks TEXT NOT NULL DEFAULT '{}', is_public INTEGER NOT NULL DEFAULT 0, likes INTEGER NOT NULL DEFAULT 0, uses INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL, updated_at DATETIME);
	CREATE TABLE cron_jobs (id TEXT PRIMARY KEY, name TEXT NOT NULL, cron_expr TEXT NOT NULL, task_type TEXT NOT NULL, payload TEXT NOT NULL DEFAULT '{}', enabled INTEGER NOT NULL DEFAULT 1, last_run_at DATETIME, last_run_status TEXT, next_run_at DATETIME, created_at DATETIME NOT NULL);
	`
	db.Exec(schema)
	return db
}


func TestPhase10_Templates(t *testing.T) {
	db := newServerDB(t)
	defer db.Close()

	s := NewServerWithDB(8080, db)

	// 1. 列出公开模板（应有 5 个预置）
	req := httptest.NewRequest("GET", "/api/v1/templates", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List: expected 200, got %d", rec.Code)
	}

	var listResp struct {
		Templates []map[string]interface{} `json:"templates"`
		Count     int                      `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &listResp)
	if listResp.Count != 5 {
		t.Errorf("Expected 5 seeded templates, got %d", listResp.Count)
	}

	// 2. 获取单个模板
	req = httptest.NewRequest("GET", "/api/v1/templates/tpl_house_001", nil)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Get template: expected 200, got %d", rec.Code)
	}
	var tpl map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &tpl)
	if tpl["name"] != "简约房屋" {
		t.Errorf("Wrong name: %v", tpl["name"])
	}

	// 3. 点赞
	req = httptest.NewRequest("POST", "/api/v1/templates/tpl_house_001/like", nil)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Like: expected 200, got %d", rec.Code)
	}
}

func TestPhase10_CronFlow(t *testing.T) {
	db := newServerDB(t)
	defer db.Close()

	s := NewServerWithDB(8080, db)

	// 1. 创建 cron job（无需认证）
	cronBody, _ := json.Marshal(map[string]interface{}{
		"name":      "test job",
		"cron_expr": "0 0 * * *",
		"task_type": "echo",
		"payload":   map[string]interface{}{"message": "hello"},
		"enabled":   true,
	})
	req := httptest.NewRequest("POST", "/api/v1/crons", bytes.NewReader(cronBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create cron: %d %s", rec.Code, rec.Body.String())
	}

	// 2. 列出 crons
	req = httptest.NewRequest("GET", "/api/v1/crons", nil)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("List crons: %d", rec.Code)
	}
	var listResp struct {
		Jobs  []map[string]interface{} `json:"jobs"`
		Count int                      `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &listResp)
	if listResp.Count != 1 {
		t.Errorf("Expected 1 cron, got %d", listResp.Count)
	}
}

func TestPhase10_NewServerWithoutDB(t *testing.T) {
	// 兼容旧用法：NewServer 不传 db 也能工作
	s := NewServer(8080)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Health: %d", rec.Code)
	}
}
