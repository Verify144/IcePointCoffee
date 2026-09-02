package task

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()
	if id1 == id2 {
		t.Error("IDs should be unique")
	}
	if len(id1) < 10 {
		t.Error("ID should be long enough")
	}
}

func TestListStore(t *testing.T) {
	store := NewListStore()

	t1 := NewTask("test", map[string]interface{}{"foo": "bar"})
	t1.UserID = "user1"
	if err := store.Create(t1); err != nil {
		t.Errorf("Create should succeed: %v", err)
	}

	// 重复创建应失败
	if err := store.Create(t1); err == nil {
		t.Error("Duplicate create should fail")
	}

	// Get
	got, err := store.Get(t1.ID)
	if err != nil {
		t.Errorf("Get should succeed: %v", err)
	}
	if got.Type != "test" {
		t.Errorf("Type mismatch: %s", got.Type)
	}

	// Update
	got.Status = StatusRunning
	store.Update(got)
	got2, _ := store.Get(t1.ID)
	if got2.Status != StatusRunning {
		t.Error("Update should work")
	}

	// List
	t2 := NewTask("other", nil)
	store.Create(t2)
	list, _ := store.List(ListFilter{})
	if len(list) != 2 {
		t.Errorf("List should return 2, got %d", len(list))
	}

	// Filter
	list2, _ := store.List(ListFilter{Status: "running"})
	if len(list2) != 1 {
		t.Errorf("Filter should return 1, got %d", len(list2))
	}

	// Count
	count, _ := store.Count()
	if count != 2 {
		t.Errorf("Count should be 2, got %d", count)
	}

	// Delete
	store.Delete(t1.ID)
	count2, _ := store.Count()
	if count2 != 1 {
		t.Errorf("Count after delete should be 1, got %d", count2)
	}
}

func TestManagerSubmit(t *testing.T) {
	store := NewListStore()
	mgr := NewManager(store, 2)
	mgr.Register("echo", func(ctx context.Context, t *Task) error {
		t.Result = "ok"
		return nil
	})
	mgr.Start()
	defer mgr.Stop()

	task := NewTask("echo", nil)
	if err := mgr.Submit(task); err != nil {
		t.Errorf("Submit should succeed: %v", err)
	}

	// 等待执行
	time.Sleep(200 * time.Millisecond)

	got, _ := mgr.Get(task.ID)
	if got.Status != StatusSuccess {
		t.Errorf("Task should be success, got %s", got.Status)
	}
	if got.Result != "ok" {
		t.Errorf("Result should be 'ok', got %s", got.Result)
	}
}

func TestManagerNoHandler(t *testing.T) {
	store := NewListStore()
	mgr := NewManager(store, 1)
	mgr.Start()
	defer mgr.Stop()

	task := NewTask("unknown", nil)
	mgr.Submit(task)
	time.Sleep(200 * time.Millisecond)

	got, _ := mgr.Get(task.ID)
	if got.Status != StatusFailed {
		t.Errorf("Unknown task should fail, got %s", got.Status)
	}
}

func TestManagerRetry(t *testing.T) {
	store := NewListStore()
	mgr := NewManager(store, 1)

	var attempts int32
	mgr.Register("flaky", func(ctx context.Context, t *Task) error {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return errors.New("transient error")
		}
		t.Result = "succeeded after retries"
		return nil
	})
	mgr.Start()
	defer mgr.Stop()

	task := NewTask("flaky", nil)
	task.MaxRetries = 5
	mgr.Submit(task)
	// 退避：第1次1s, 第2次2s, 第3次成功 -> 总共3s+执行时间
	time.Sleep(5 * time.Second)

	got, _ := mgr.Get(task.ID)
	if got.Status != StatusSuccess {
		t.Errorf("Task should succeed after retries, got %s (err: %s)", got.Status, got.Error)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("Should be 3 attempts, got %d", attempts)
	}
	if got.Retries != 2 {
		t.Errorf("Should record 2 retries, got %d", got.Retries)
	}
}

func TestManagerCancel(t *testing.T) {
	store := NewListStore()
	mgr := NewManager(store, 1)

	mgr.Register("slow", func(ctx context.Context, t *Task) error {
		select {
		case <-time.After(2 * time.Second):
			return nil
		case <-ctx.Done():
			return ErrCancelled
		}
	})
	mgr.Start()
	defer mgr.Stop()

	task := NewTask("slow", nil)
	mgr.Submit(task)
	time.Sleep(100 * time.Millisecond)

	// 取消
	mgr.Cancel(task.ID)
	time.Sleep(500 * time.Millisecond)

	got, _ := mgr.Get(task.ID)
	if got.Status != StatusCancelled {
		t.Errorf("Task should be cancelled, got %s", got.Status)
	}
}

func TestManagerConcurrent(t *testing.T) {
	store := NewListStore()
	mgr := NewManager(store, 4)

	var counter int32
	mgr.Register("counter", func(ctx context.Context, t *Task) error {
		atomic.AddInt32(&counter, 1)
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	mgr.Start()
	defer mgr.Stop()

	// 提交 10 个任务
	for i := 0; i < 10; i++ {
		task := NewTask("counter", nil)
		mgr.Submit(task)
	}

	// 等待
	time.Sleep(2 * time.Second)

	if atomic.LoadInt32(&counter) != 10 {
		t.Errorf("All tasks should complete, got %d", counter)
	}
}

func TestUpdateProgress(t *testing.T) {
	store := NewListStore()
	mgr := NewManager(store, 1)

	mgr.Register("progress", func(ctx context.Context, t *Task) error {
		for i := 0; i <= 100; i += 25 {
			mgr.UpdateProgress(t.ID, i, "")
		}
		return nil
	})
	mgr.Start()
	defer mgr.Stop()

	task := NewTask("progress", nil)
	mgr.Submit(task)
	time.Sleep(200 * time.Millisecond)

	got, _ := mgr.Get(task.ID)
	if got.Progress != 100 {
		t.Errorf("Progress should be 100, got %d", got.Progress)
	}
}

func TestNewTask(t *testing.T) {
	t1 := NewTask("test", map[string]interface{}{"a": 1})
	if t1.ID == "" {
		t.Error("ID should be set")
	}
	if t1.Status != StatusPending {
		t.Error("Status should be pending")
	}
	if t1.Priority != PriorityNormal {
		t.Error("Priority should be normal")
	}
	if t1.MaxRetries != 3 {
		t.Error("MaxRetries should be 3")
	}
	if t1.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}
