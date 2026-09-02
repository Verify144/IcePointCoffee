package db

import (
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"
)

func newCmdHistoryDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, _ := sql.Open("sqlite", filepath.Join(dir, "cmd.db")+"?_journal=WAL")
	db.SetMaxOpenConns(1)
	schema := `
	CREATE TABLE command_history (
		id TEXT PRIMARY KEY, tool TEXT NOT NULL, command TEXT NOT NULL,
		args TEXT NOT NULL DEFAULT '{}', output TEXT NOT NULL DEFAULT '',
		success INTEGER NOT NULL DEFAULT 1, dangerous INTEGER NOT NULL DEFAULT 0,
		duration_ms INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL
	);
	CREATE INDEX idx_history_created ON command_history(created_at DESC);
	CREATE INDEX idx_history_tool ON command_history(tool);
	`
	db.Exec(schema)
	return db
}

func TestCmdHistoryAdd(t *testing.T) {
	db := newCmdHistoryDB(t)
	defer db.Close()
	store := NewCommandHistoryStore(db)

	h := &CommandHistory{
		Tool: "mc_command", Command: "time set day",
		Output: "Time set to day", Success: true, Dangerous: false,
		DurationMs: 50, CreatedAt: time.Now(),
	}
	if err := store.Add(h); err != nil {
		t.Fatal(err)
	}
	if h.ID == "" {
		t.Error("ID should be set")
	}
}

func TestCmdHistoryList(t *testing.T) {
	db := newCmdHistoryDB(t)
	defer db.Close()
	store := NewCommandHistoryStore(db)

	for i := 0; i < 5; i++ {
		store.Add(&CommandHistory{
			Tool: "mc_command", Command: "test", Success: true,
			DurationMs: 10, CreatedAt: time.Now(),
		})
	}

	list, _ := store.List(10)
	if len(list) != 5 {
		t.Errorf("Expected 5, got %d", len(list))
	}
}

func TestCmdHistoryClear(t *testing.T) {
	db := newCmdHistoryDB(t)
	defer db.Close()
	store := NewCommandHistoryStore(db)
	store.Add(&CommandHistory{Tool: "x", Command: "y", CreatedAt: time.Now()})
	store.Add(&CommandHistory{Tool: "x", Command: "y", CreatedAt: time.Now()})

	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	n, _ := store.Count()
	if n != 0 {
		t.Errorf("Expected 0 after clear, got %d", n)
	}
}
