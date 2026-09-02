// Smoke test binary for Phase 10
// Run: go run ./cmd/smoke
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"

	"github.com/Verify144/IcePointCoffee/internal/server"
)

func main() {
	dir, _ := filepath.Abs("/tmp/icepoint_smoke")
	dbPath := filepath.Join(dir, "smoke.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal=WAL")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, username TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', created_at DATETIME NOT NULL, updated_at DATETIME);
	CREATE TABLE IF NOT EXISTS api_tokens (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, token_hash TEXT NOT NULL UNIQUE, scopes TEXT NOT NULL DEFAULT '[]', expires_at DATETIME, last_used_at DATETIME, created_at DATETIME NOT NULL);
	CREATE TABLE IF NOT EXISTS build_templates (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, description TEXT, category TEXT NOT NULL DEFAULT 'custom', params_schema TEXT NOT NULL DEFAULT '{}', blocks TEXT NOT NULL DEFAULT '{}', is_public INTEGER NOT NULL DEFAULT 0, likes INTEGER NOT NULL DEFAULT 0, uses INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL, updated_at DATETIME);
	CREATE TABLE IF NOT EXISTS cron_jobs (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, cron_expr TEXT NOT NULL, task_type TEXT NOT NULL, payload TEXT NOT NULL DEFAULT '{}', enabled INTEGER NOT NULL DEFAULT 1, last_run_at DATETIME, last_run_status TEXT, next_run_at DATETIME, created_at DATETIME NOT NULL);
	`
	db.Exec(schema)

	s := server.NewServerWithDB(8765, db)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
	defer s.Stop()

	// Wait for server
	time.Sleep(500 * time.Millisecond)
	fmt.Println("Server started on :8765")

	// 冒烟测试在外部 curl
	select {}
}
