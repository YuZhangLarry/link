package handler

import (
	"context"
	"log"
	"sync"
	"time"
)

// TaskProcessor 异步任务处理器
type TaskProcessor struct {
	queue            chan *Task
	workerCount      int
	wg               sync.WaitGroup
	knowledgeHandler *KnowledgeHandlerFull
}

// Task 异步任务
type Task struct {
	ID        string
	Type      string
	Data      interface{}
	CreatedAt time.Time
}

// KnowledgeTaskData 知识库处理任务数据
type KnowledgeTaskData struct {
	KnowledgeID  string
	TenantID     int64
	UserID       int64
	KBID         string
	ChunkSize    int
	ChunkOverlap int
}

// NewTaskProcessor 创建任务处理器
func NewTaskProcessor(handler *KnowledgeHandlerFull, workerCount int) *TaskProcessor {
	return &TaskProcessor{
		queue:            make(chan *Task, 100),
		workerCount:      workerCount,
		knowledgeHandler: handler,
	}
}

// Start 启动任务处理器
func (p *TaskProcessor) Start() {
	log.Printf("[TaskProcessor] Starting with %d workers", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	log.Println("[TaskProcessor] Task processor started")
}

// Stop 停止任务处理器
func (p *TaskProcessor) Stop() {
	log.Println("[TaskProcessor] Stopping task processor...")
	close(p.queue)
	p.wg.Wait()
	log.Println("[TaskProcessor] Task processor stopped")
}

// Submit 提交任务
func (p *TaskProcessor) Submit(task *Task) error {
	select {
	case p.queue <- task:
		log.Printf("[TaskProcessor] Task submitted: type=%s, id=%s", task.Type, task.ID)
		return nil
	case <-time.After(5 * time.Second):
		return ErrQueueFull
	}
}

// worker 工作协程
func (p *TaskProcessor) worker(id int) {
	defer p.wg.Done()

	log.Printf("[TaskProcessor-%d] Worker started", id)

	for task := range p.queue {
		p.processTask(id, task)
	}

	log.Printf("[TaskProcessor-%d] Worker stopped", id)
}

// processTask 处理单个任务
func (p *TaskProcessor) processTask(workerID int, task *Task) {
	log.Printf("[TaskProcessor-%d] Processing task: type=%s, id=%s", workerID, task.Type, task.ID)

	startTime := time.Now()

	var err error
	switch task.Type {
	case TaskTypeKnowledgeProcess:
		err = p.processKnowledgeTask(task)
	case TaskTypeGraphBuild:
		err = p.processGraphBuildTask(task)
	default:
		log.Printf("[TaskProcessor-%d] Unknown task type: %s", workerID, task.Type)
	}

	duration := time.Since(startTime)
	if err != nil {
		log.Printf("[TaskProcessor-%d] Task failed: type=%s, id=%s, error=%v, duration=%v",
			workerID, task.Type, task.ID, err, duration)
	} else {
		log.Printf("[TaskProcessor-%d] Task completed: type=%s, id=%s, duration=%v",
			workerID, task.Type, task.ID, duration)
	}
}

// processKnowledgeTask 处理知识库任务
func (p *TaskProcessor) processKnowledgeTask(task *Task) error {
	data, ok := task.Data.(*KnowledgeTaskData)
	if !ok {
		return ErrInvalidTaskData
	}

	// 使用更长的超时时间（30分钟），因为文档处理、embedding生成和Milvus插入可能需要较长时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	return p.knowledgeHandler.ProcessKnowledgeTask(ctx, data.KnowledgeID, data.ChunkSize, data.ChunkOverlap)
}

// processGraphBuildTask 处理图谱构建任务
func (p *TaskProcessor) processGraphBuildTask(task *Task) error {
	// TODO: 实现图谱构建任务
	log.Printf("[TaskProcessor] Graph build task: %v", task)
	return nil
}

// Task types
const (
	TaskTypeKnowledgeProcess = "knowledge_process"
	TaskTypeGraphBuild       = "graph_build"
)

// Errors
var (
	ErrQueueFull       = &TaskError{Code: "QUEUE_FULL", Message: "Task queue is full"}
	ErrInvalidTaskData = &TaskError{Code: "INVALID_DATA", Message: "Invalid task data"}
)

// TaskError 任务错误
type TaskError struct {
	Code    string
	Message string
}

func (e *TaskError) Error() string {
	return e.Message
}

// Stats 任务处理器统计
type Stats struct {
	QueueLength   int
	ActiveWorkers int
}

// GetStats 获取统计信息
func (p *TaskProcessor) GetStats() *Stats {
	return &Stats{
		QueueLength:   len(p.queue),
		ActiveWorkers: p.workerCount,
	}
}
