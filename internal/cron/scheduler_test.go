package cron

import (
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "modernc.org/sqlite"
)

func newCronDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, _ := sql.Open("sqlite", filepath.Join(dir, "cron.db")+"?_journal=WAL")
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE cron_jobs (id TEXT PRIMARY KEY, name TEXT NOT NULL, cron_expr TEXT NOT NULL, task_type TEXT NOT NULL, payload TEXT NOT NULL DEFAULT '{}', enabled INTEGER NOT NULL DEFAULT 1, last_run_at DATETIME, last_run_status TEXT, next_run_at DATETIME, created_at DATETIME NOT NULL);
	`
	db.Exec(schema)
	return db
}

func TestCronCreate(t *testing.T) {
	db := newCronDB(t)
	defer db.Close()

	store := NewStore(db)
	job := &CronJob{
		ID:        "test_job_1",
		Name:      "Daily cleanup",
		CronExpr:  "0 0 * * *",
		TaskType:  "echo",
		Payload:   map[string]interface{}{"msg": "hello"},
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	if err := store.Create(job); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("test_job_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Daily cleanup" || got.CronExpr != "0 0 * * *" {
		t.Errorf("Mismatch: %+v", got)
	}
	if got.Payload["msg"] != "hello" {
		t.Error("Payload mismatch")
	}
}

func TestCronList(t *testing.T) {
	db := newCronDB(t)
	defer db.Close()

	store := NewStore(db)
	for i := 0; i < 3; i++ {
		store.Create(&CronJob{
			Enabled: true, CreatedAt: time.Now(),
		})
	}

	jobs, _ := store.List(false)
	if len(jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(jobs))
	}

	enabled, _ := store.List(true)
	if len(enabled) != 3 {
		t.Errorf("Expected 3 enabled, got %d", len(enabled))
	}
}

func TestCronDisable(t *testing.T) {
	db := newCronDB(t)
	defer db.Close()

	store := NewStore(db)
	store.Create(&CronJob{
		ID: "d1", Enabled: true, CreatedAt: time.Now(),
	})

	// 禁用
	job, _ := store.Get("d1")
	job.Enabled = false
	store.Update(job)

	// 只列出启用的应该看不到
	jobs, _ := store.List(true)
	if len(jobs) != 0 {
		t.Errorf("Expected 0 enabled jobs, got %d", len(jobs))
	}
}

func TestCronDelete(t *testing.T) {
	db := newCronDB(t)
	defer db.Close()

	store := NewStore(db)
	store.Create(&CronJob{
		ID: "x1", Enabled: true, CreatedAt: time.Now(),
	})

	if err := store.Delete("x1"); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("x1"); err == nil {
		t.Error("Should fail on second delete")
	}
}

func TestCronDue(t *testing.T) {
	db := newCronDB(t)
	defer db.Close()

	store := NewStore(db)
	past := time.Now().Add(-1 * time.Hour)
	store.Create(&CronJob{
		Enabled: true, CreatedAt: time.Now(), NextRunAt: &past,
	})

	due, _ := store.GetDueJobs(time.Now())
	if len(due) != 1 {
		t.Errorf("Expected 1 due job, got %d", len(due))
	}
}

func TestMatchesCron(t *testing.T) {
	now := time.Date(2026, 9, 2, 14, 30, 0, 0, time.UTC)
	if !matchesCron("* * * * *", now) {
		t.Error("Should match wildcard")
	}
	if !matchesCron("30 14 * * *", now) {
		t.Error("Should match exact 30 14")
	}
	if matchesCron("0 0 * * *", now) {
		t.Error("Should not match 0 0 at 14:30")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	db := newCronDB(t)
	defer db.Close()

	store := NewStore(db)
	sched := NewScheduler(store, nil)
	sched.Start()
	time.Sleep(50 * time.Millisecond)
	sched.Stop()
	// 不能再次 stop，但要安全
	sched.Stop()
}

func TestParseCronNext(t *testing.T) {
	now := time.Date(2026, 9, 2, 14, 30, 0, 0, time.UTC)
	next := parseCronNext("* * * * *", now)
	if next.Before(now) {
		t.Error("Next should be after now")
	}
	if next.Sub(now) > 1*time.Hour {
		t.Errorf("Next should be within 1 hour, got %v", next.Sub(now))
	}
}
