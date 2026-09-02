// Package cron 提供定时任务调度。
package cron

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/task"
)

// CronJob 定时任务
type CronJob struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Name          string     `json:"name"`
	CronExpr      string     `json:"cron_expr"` // 标准 cron 表达式
	TaskType      string     `json:"task_type"`  // 对应 task handler 类型
	Payload       map[string]interface{} `json:"payload"`
	Enabled       bool       `json:"enabled"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	LastRunStatus string     `json:"last_run_status,omitempty"` // success/failed
	NextRunAt     *time.Time `json:"next_run_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Store Cron 存储
type Store struct {
	db *sql.DB
}

// NewStore 创建 cron 存储
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create 创建 cron job
func (s *Store) Create(j *CronJob) error {
	if j.ID == "" {
		j.ID = generateCronID()
	}
	payloadJSON, _ := json.Marshal(j.Payload)

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO cron_jobs (id, user_id, name, cron_expr, task_type, payload, enabled, created_at)
		 VALUES (?,?,?,?,?,?,?,?)`,
		j.ID, j.UserID, j.Name, j.CronExpr, j.TaskType, string(payloadJSON), boolToInt(j.Enabled), j.CreatedAt)
	return err
}

// Get 获取 job
func (s *Store) Get(id string) (*CronJob, error) {
	var j CronJob
	var payloadJSON string
	var lastRunAt, nextRunAt sql.NullTime
	var lastStatus sql.NullString

	err := s.db.QueryRowContext(context.Background(),
		`SELECT id, user_id, name, cron_expr, task_type, payload, enabled, last_run_at, last_run_status, next_run_at, created_at
		 FROM cron_jobs WHERE id=?`, id).
		Scan(&j.ID, &j.UserID, &j.Name, &j.CronExpr, &j.TaskType, &payloadJSON,
			&j.Enabled, &lastRunAt, &lastStatus, &nextRunAt, &j.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cron job not found")
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(payloadJSON), &j.Payload)
	if lastRunAt.Valid {
		j.LastRunAt = &lastRunAt.Time
	}
	if nextRunAt.Valid {
		j.NextRunAt = &nextRunAt.Time
	}
	if lastStatus.Valid {
		j.LastRunStatus = lastStatus.String
	}

	return &j, nil
}

// List 列出 jobs
func (s *Store) List(userID string, enabledOnly bool) ([]*CronJob, error) {
	query := `SELECT id, user_id, name, cron_expr, task_type, payload, enabled, last_run_at, last_run_status, next_run_at, created_at
	          FROM cron_jobs WHERE 1=1`
	var args []interface{}

	if userID != "" {
		query += " AND user_id=?"
		args = append(args, userID)
	}
	if enabledOnly {
		query += " AND enabled=1"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*CronJob
	for rows.Next() {
		var j CronJob
		var payloadJSON string
		var lastRunAt, nextRunAt sql.NullTime
		var lastStatus sql.NullString
		if err := rows.Scan(&j.ID, &j.UserID, &j.Name, &j.CronExpr, &j.TaskType, &payloadJSON,
			&j.Enabled, &lastRunAt, &lastStatus, &nextRunAt, &j.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(payloadJSON), &j.Payload)
		if lastRunAt.Valid {
			j.LastRunAt = &lastRunAt.Time
		}
		if nextRunAt.Valid {
			j.NextRunAt = &nextRunAt.Time
		}
		if lastStatus.Valid {
			j.LastRunStatus = lastStatus.String
		}
		jobs = append(jobs, &j)
	}
	return jobs, rows.Err()
}

// Update 更新 job
func (s *Store) Update(j *CronJob) error {
	payloadJSON, _ := json.Marshal(j.Payload)
	_, err := s.db.ExecContext(context.Background(),
		`UPDATE cron_jobs SET name=?, cron_expr=?, task_type=?, payload=?, enabled=?, last_run_at=?, last_run_status=?, next_run_at=?
		 WHERE id=?`,
		j.Name, j.CronExpr, j.TaskType, string(payloadJSON), boolToInt(j.Enabled),
		j.LastRunAt, j.LastRunStatus, j.NextRunAt, j.ID)
	return err
}

// Delete 删除 job
func (s *Store) Delete(id, userID string) error {
	res, err := s.db.ExecContext(context.Background(),
		`DELETE FROM cron_jobs WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cron job not found or not owned")
	}
	return nil
}

// UpdateRunStatus 更新运行状态
func (s *Store) UpdateRunStatus(id string, status string, nextRun *time.Time) error {
	now := time.Now()
	_, err := s.db.ExecContext(context.Background(),
		`UPDATE cron_jobs SET last_run_at=?, last_run_status=?, next_run_at=? WHERE id=?`,
		now, status, nextRun, id)
	return err
}

// GetDueJobs 获取即将到期的 jobs
func (s *Store) GetDueJobs(before time.Time) ([]*CronJob, error) {
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT id, user_id, name, cron_expr, task_type, payload, enabled, last_run_at, last_run_status, next_run_at, created_at
		 FROM cron_jobs WHERE enabled=1 AND (next_run_at IS NULL OR next_run_at<=?)
		 ORDER BY next_run_at ASC LIMIT 100`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*CronJob
	for rows.Next() {
		var j CronJob
		var payloadJSON string
		var lastRunAt, nextRunAt sql.NullTime
		var lastStatus sql.NullString
		rows.Scan(&j.ID, &j.UserID, &j.Name, &j.CronExpr, &j.TaskType, &payloadJSON,
			&j.Enabled, &lastRunAt, &lastStatus, &nextRunAt, &j.CreatedAt)
		json.Unmarshal([]byte(payloadJSON), &j.Payload)
		if lastRunAt.Valid {
			j.LastRunAt = &lastRunAt.Time
		}
		if nextRunAt.Valid {
			j.NextRunAt = &nextRunAt.Time
		}
		if lastStatus.Valid {
			j.LastRunStatus = lastStatus.String
		}
		jobs = append(jobs, &j)
	}
	return jobs, rows.Err()
}

// ==== 调度器 ====

// Scheduler 定时任务调度器
type Scheduler struct {
	store       *Store
	taskManager *task.Manager
	mu          sync.Mutex
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
}

// NewScheduler 创建调度器
func NewScheduler(store *Store, taskMgr *task.Manager) *Scheduler {
	return &Scheduler{
		store:       store,
		taskManager: taskMgr,
		stopCh:      make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
	log.Println("[Cron] Scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()
	s.wg.Wait()
	log.Println("[Cron] Scheduler stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 立即运行一次
	s.tick()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Scheduler) tick() {
	ctx := context.Background()
	jobs, err := s.store.GetDueJobs(time.Now())
	if err != nil {
		log.Printf("[Cron] GetDueJobs error: %v", err)
		return
	}

	for _, job := range jobs {
		s.executeJob(ctx, job)
	}
}

func (s *Scheduler) executeJob(ctx context.Context, job *CronJob) {
	log.Printf("[Cron] Running job: %s (%s)", job.Name, job.ID)

	// 计算下次运行时间
	nextRun := parseCronNext(job.CronExpr, time.Now())

	// 构造任务
	t := task.NewTask(job.TaskType, job.Payload)
	t.UserID = job.UserID

	// 提交到任务管理器
	if err := s.taskManager.Submit(t); err != nil {
		s.store.UpdateRunStatus(job.ID, "failed", &nextRun)
		log.Printf("[Cron] Job %s failed to submit: %v", job.ID, err)
		return
	}

	s.store.UpdateRunStatus(job.ID, "success", &nextRun)
	log.Printf("[Cron] Job %s completed, next run: %v", job.ID, nextRun)
}

// ==== 工具函数 ====

func generateCronID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("cron_%x", b)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// parseCronNext 解析 cron 表达式计算下次运行时间（简化实现，支持标准格式）
// 支持: <sec> <min> <hour> <day> <month> <weekday>
// 这里实现简化版：只支持 5 字段 (min hour day month weekday)
func parseCronNext(expr string, from time.Time) time.Time {
	// 简化：解析 "*/n * * * *" 格式
	// 计算未来最近的匹配时间
	next := from.Add(1 * time.Minute)

	// 完整实现建议用 robfig/cron 库
	// 此处为占位，确保调度器能正常运行
	for i := 0; i < 60*24*31; i++ { // 最多遍历31天
		if matchesCron(expr, next) {
			return next
		}
		next = next.Add(1 * time.Minute)
	}
	return next.Add(1 * time.Hour) // 默认1小时后
}

func matchesCron(expr string, t time.Time) bool {
	fields := splitFields(expr)
	if len(fields) < 5 {
		return false
	}

	return matchesField(fields[0], t.Minute(), 0, 59) &&
		matchesField(fields[1], t.Hour(), 0, 23) &&
		matchesField(fields[2], t.Day(), 1, 31) &&
		matchesField(fields[3], int(t.Month()), 1, 12) &&
		matchesField(fields[4], int(t.Weekday()), 0, 6)
}

func splitFields(expr string) []string {
	// 分割标准 5 字段格式（分 时 日 月 周）
	fields := make([]string, 0, 6)
	for _, f := range strings.Fields(expr) {
		if f != "" {
			fields = append(fields, f)
		}
	}
	return fields
}

func matchesField(field string, value, min, max int) bool {
	if field == "*" {
		return true
	}
	if strings.HasPrefix(field, "*/") {
		step := 0
		fmt.Sscanf(field[2:], "%d", &step)
		if step > 0 && value%step == 0 {
			return true
		}
	}
	var v int
	if n, _ := fmt.Sscanf(field, "%d", &v); n == 1 {
		return v == value
	}
	return false
}
