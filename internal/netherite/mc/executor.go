// Package mc 提供 Minecraft Bedrock 连接管理。
// 整合 Raknet 传输层、FB 认证、登录握手和游戏协议。
package mc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
)

// CommandTask 单个命令任务。
type CommandTask struct {
	ID        string
	Command   string
	Priority  int
	CreatedAt time.Time
	StartAt   time.Time
	DoneAt    time.Time
	Result    *CommandOutput
	Error     error
	Status    CommandStatus
}

// CommandStatus 命令状态。
type CommandStatus uint8

const (
	CommandStatusPending CommandStatus = iota
	CommandStatusRunning
	CommandStatusDone
	CommandStatusFailed
)

// CommandOutput 命令输出。
type CommandOutput struct {
	SuccessCount int32
	Messages    []string
	Raw         *protocol.CommandOutput
}

// AsyncExecutor 异步命令执行器。
type AsyncExecutor struct {
	client       *Client
	taskQueue    chan *CommandTask
	resultQueue  chan *CommandTask
	workerCount  int
	maxRetries   int
	rateLimit    time.Duration

	// 状态
	running      atomic.Bool
	processed    atomic.Int64
	failed       atomic.Int64
	wg           sync.WaitGroup
	stopChan     chan struct{}

	// 事件
	onCommandSent  func(*CommandTask)
	onCommandDone func(*CommandTask)
}

// NewAsyncExecutor 创建异步执行器。
func NewAsyncExecutor(client *Client, workers int) *AsyncExecutor {
	if workers <= 0 {
		workers = 4
	}
	ex := &AsyncExecutor{
		client:      client,
		taskQueue:   make(chan *CommandTask, 1024),
		resultQueue: make(chan *CommandTask, 256),
		workerCount: workers,
		maxRetries:  3,
		rateLimit:   50 * time.Millisecond,
		stopChan:    make(chan struct{}),
	}
	return ex
}

// Start 启动执行器。
func (e *AsyncExecutor) Start() {
	if !e.running.CompareAndSwap(false, true) {
		return
	}

	for i := 0; i < e.workerCount; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
	log.Printf("[Executor] Started %d workers", e.workerCount)
}

// Stop 停止执行器。
func (e *AsyncExecutor) Stop() {
	if !e.running.CompareAndSwap(true, false) {
		return
	}

	close(e.stopChan)
	e.wg.Wait()
	close(e.taskQueue)
	close(e.resultQueue)
	log.Printf("[Executor] Stopped, processed=%d failed=%d", e.processed.Load(), e.failed.Load())
}

// worker 执行器工作协程。
func (e *AsyncExecutor) worker(id int) {
	defer e.wg.Done()

	for {
		select {
		case <-e.stopChan:
			return
		case task, ok := <-e.taskQueue:
			if !ok {
				return
			}
			e.executeTask(task)
		}
	}
}

// executeTask 执行单个任务。
func (e *AsyncExecutor) executeTask(task *CommandTask) {
	task.Status = CommandStatusRunning
	task.StartAt = time.Now()

	if e.onCommandSent != nil {
		e.onCommandSent(task)
	}

	// 最多重试 maxRetries 次
	var lastErr error
	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			log.Printf("[Executor] Retrying command (attempt %d/%d): %s", attempt+1, e.maxRetries+1, task.Command)
		}

		output, err := e.client.SendCommand(context.Background(), task.Command)
		if err != nil {
			lastErr = err
			continue
		}

		// 成功
		task.Result = &CommandOutput{
			SuccessCount: output.SuccessCount,
			Messages:    []string{},
			Raw:         output,
		}
		for _, msg := range output.Messages {
			task.Result.Messages = append(task.Result.Messages, msg.Message)
		}

		task.Status = CommandStatusDone
		task.DoneAt = time.Now()
		e.processed.Add(1)

		select {
		case e.resultQueue <- task:
		default:
		}

		if e.onCommandDone != nil {
			e.onCommandDone(task)
		}

		// 速率限制
		time.Sleep(e.rateLimit)
		return
	}

	// 全部失败
	task.Error = lastErr
	task.Status = CommandStatusFailed
	task.DoneAt = time.Now()
	e.failed.Add(1)

	select {
	case e.resultQueue <- task:
	default:
	}

	if e.onCommandDone != nil {
		e.onCommandDone(task)
	}
}

// Submit 提交命令。
func (e *AsyncExecutor) Submit(cmd string) *CommandTask {
	if !e.running.Load() {
		return nil
	}

	task := &CommandTask{
		ID:        fmt.Sprintf("cmd_%d", time.Now().UnixNano()),
		Command:   cmd,
		Priority:  0,
		CreatedAt: time.Now(),
		Status:    CommandStatusPending,
	}

	select {
	case e.taskQueue <- task:
		return task
	default:
		log.Printf("[Executor] Task queue full, dropping task: %s", cmd)
		return nil
	}
}

// SubmitBatch 批量提交。
func (e *AsyncExecutor) SubmitBatch(cmds []string) []*CommandTask {
	var tasks []*CommandTask
	for _, cmd := range cmds {
		if task := e.Submit(cmd); task != nil {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// Result 返回已完成的结果（非阻塞）。
func (e *AsyncExecutor) Result() (*CommandTask, bool) {
	select {
	case task := <-e.resultQueue:
		return task, true
	default:
		return nil, false
	}
}

// Results 返回所有已完成的结果（非阻塞）。
func (e *AsyncExecutor) Results() []*CommandTask {
	var results []*CommandTask
	for {
		task, ok := e.Result()
		if !ok {
			break
		}
		results = append(results, task)
	}
	return results
}

// SetRateLimit 设置速率限制。
func (e *AsyncExecutor) SetRateLimit(d time.Duration) {
	e.rateLimit = d
}

// SetMaxRetries 设置最大重试次数。
func (e *AsyncExecutor) SetMaxRetries(n int) {
	e.maxRetries = n
}

// OnCommandSent 设置命令发送回调。
func (e *AsyncExecutor) OnCommandSent(fn func(*CommandTask)) {
	e.onCommandSent = fn
}

// OnCommandDone 设置命令完成回调。
func (e *AsyncExecutor) OnCommandDone(fn func(*CommandTask)) {
	e.onCommandDone = fn
}

// Stats 返回执行统计。
func (e *AsyncExecutor) Stats() ExecutorStats {
	return ExecutorStats{
		Running:   e.running.Load(),
		Processed: e.processed.Load(),
		Failed:    e.failed.Load(),
		QueueLen:  len(e.taskQueue),
		ResultLen: len(e.resultQueue),
	}
}

// ExecutorStats 执行统计。
type ExecutorStats struct {
	Running   bool
	Processed int64
	Failed    int64
	QueueLen  int
	ResultLen int
}
