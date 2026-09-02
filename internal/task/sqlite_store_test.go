package task

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal=WAL&_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	return db
}

func TestSQLiteStoreCreateAndGet(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)

	task := &Task{
		ID:        GenerateID(),
		Type:      "build",
		Payload:   map[string]interface{}{"x": 10, "y": 64, "z": -5},
		Status:    StatusPending,
		Priority:  PriorityHigh,
		UserID:    "user_1",
		Tags:      []string{"urgent", "vip"},
		CreatedAt: time.Now(),
	}

	if err := store.Create(task); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := store.Get(task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.ID != task.ID || got.Type != "build" || got.Priority != PriorityHigh {
		t.Errorf("Task mismatch: got %+v", got)
	}
	if got.Payload["x"] != float64(10) {
		t.Errorf("Payload mismatch: %+v", got.Payload)
	}
	if len(got.Tags) != 2 {
		t.Errorf("Tags mismatch: %+v", got.Tags)
	}
}

func TestSQLiteStoreUpdate(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)
	task := NewTask("echo", map[string]interface{}{"msg": "hi"})
	store.Create(task)

	task.Status = StatusSuccess
	task.Progress = 100
	task.Result = "Echo: hi"
	now := time.Now()
	task.CompletedAt = &now

	if err := store.Update(task); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ := store.Get(task.ID)
	if got.Status != StatusSuccess || got.Progress != 100 {
		t.Errorf("Update didn't persist: %+v", got)
	}
}

func TestSQLiteStoreList(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)

	// 插入 5 个不同状态的 task
	for i := 0; i < 3; i++ {
		task := NewTask("build", nil)
		task.Status = StatusSuccess
		store.Create(task)
	}
	for i := 0; i < 2; i++ {
		task := NewTask("command", nil)
		task.Status = StatusFailed
		store.Create(task)
	}

	all, _ := store.ListAll(100)
	if len(all) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(all))
	}

	success, _ := store.List(ListFilter{Status: "success"})
	if len(success) != 3 {
		t.Errorf("Expected 3 success tasks, got %d", len(success))
	}

	failed, _ := store.List(ListFilter{Status: "failed"})
	if len(failed) != 2 {
		t.Errorf("Expected 2 failed tasks, got %d", len(failed))
	}

	build, _ := store.List(ListFilter{Type: "build"})
	if len(build) != 3 {
		t.Errorf("Expected 3 build tasks, got %d", len(build))
	}
}

func TestSQLiteStoreStats(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)
	store.Create(NewTask("a", nil))
	t2 := NewTask("b", nil)
	t2.Status = StatusSuccess
	store.Create(t2)
	t3 := NewTask("c", nil)
	t3.Status = StatusFailed
	store.Create(t3)

	stats, err := store.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats["total"] != 3 {
		t.Errorf("Expected 3 total, got %d", stats["total"])
	}
	if stats["success"] != 1 {
		t.Errorf("Expected 1 success, got %d", stats["success"])
	}
	if stats["failed"] != 1 {
		t.Errorf("Expected 1 failed, got %d", stats["failed"])
	}
}

func TestSQLiteStoreDelete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)
	task := NewTask("echo", nil)
	store.Create(task)

	if err := store.Delete(task.ID); err != nil {
		t.Fatal(err)
	}

	_, err := store.Get(task.ID)
	if err == nil {
		t.Error("Task should not exist after delete")
	}
}

func TestManagerWithSQLite(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	store := NewSQLiteStore(db)
	mgr := NewManager(store, 2)
	mgr.Register("persist", func(ctx context.Context, t *Task) error {
		t.Result = "persisted"
		return nil
	})
	mgr.Start()
	defer mgr.Stop()

	task := NewTask("persist", map[string]interface{}{"x": 1})
	if err := mgr.Submit(task); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	// 创建新的 Manager 和 SQLite store，验证持久化
	store2 := NewSQLiteStore(db)
	got, err := store2.Get(task.ID)
	if err != nil {
		t.Fatalf("Persistence failed: %v", err)
	}
	if got.Status != StatusSuccess || got.Result != "persisted" {
		t.Errorf("Task should be persisted, got %+v", got)
	}
}
