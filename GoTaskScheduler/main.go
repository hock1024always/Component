package main

import (
	"fmt"
	"sync"
	"time"
)

// Task 任务结构体
type Task struct {
	ID          string
	Name        string
	ClusterID   string
	Status      string // pending, running, completed, failed
	Priority    int    // 1-10, 越高优先级越大
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	WorkerID    string
}

// Worker 工作节点结构体
type Worker struct {
	ID        string
	ClusterID string
	Status    string // idle, busy
	Capacity  int    // 同时处理任务数
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	tasks       map[string]*Task
	workers     map[string]*Worker
	clusters    map[string][]string // clusterID -> workerIDs
	taskQueue   chan *Task
	workerMutex sync.RWMutex
	taskMutex   sync.RWMutex
	stopChan    chan bool
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks:     make(map[string]*Task),
		workers:   make(map[string]*Worker),
		clusters:  make(map[string][]string),
		taskQueue: make(chan *Task, 100),
		stopChan:  make(chan bool),
	}
}

// AddWorker 添加工作节点
func (ts *TaskScheduler) AddWorker(worker *Worker) {
	ts.workerMutex.Lock()
	defer ts.workerMutex.Unlock()

	ts.workers[worker.ID] = worker
	ts.clusters[worker.ClusterID] = append(ts.clusters[worker.ClusterID], worker.ID)
	fmt.Printf("添加工作节点: %s (集群: %s)\n", worker.ID, worker.ClusterID)
}

// SubmitTask 提交任务
func (ts *TaskScheduler) SubmitTask(task *Task) {
	ts.taskMutex.Lock()
	task.Status = "pending"
	task.CreatedAt = time.Now()
	ts.tasks[task.ID] = task
	ts.taskMutex.Unlock()

	select {
	case ts.taskQueue <- task:
		fmt.Printf("任务已提交: %s\n", task.ID)
	default:
		fmt.Printf("任务队列已满，稍后重试: %s\n", task.ID)
	}
}

// Schedule 调度任务到工作节点
func (ts *TaskScheduler) Schedule(task *Task) bool {
	ts.workerMutex.RLock()
	defer ts.workerMutex.RUnlock()

	// 优先在本集群内寻找空闲工作节点
	clusterWorkers := ts.clusters[task.ClusterID]
	for _, workerID := range clusterWorkers {
		worker := ts.workers[workerID]
		if worker.Status == "idle" {
			return ts.assignTask(task, worker)
		}
	}

	// 如果本集群没有空闲节点，在其他集群寻找
	for clusterID, workerIDs := range ts.clusters {
		if clusterID == task.ClusterID {
			continue
		}
		for _, workerID := range workerIDs {
			worker := ts.workers[workerID]
			if worker.Status == "idle" {
				return ts.assignTask(task, worker)
			}
		}
	}

	return false // 没有找到合适的worker
}

// assignTask 分配任务给工作节点
func (ts *TaskScheduler) assignTask(task *Task, worker *Worker) bool {
	ts.workerMutex.RUnlock()
	ts.workerMutex.Lock()
	defer ts.workerMutex.Unlock()

	// 双重检查worker状态
	if worker.Status != "idle" {
		return false
	}

	// 分配任务
	worker.Status = "busy"
	now := time.Now()
	task.Status = "running"
	task.StartedAt = &now
	task.WorkerID = worker.ID

	fmt.Printf("任务 %s 已分配给工作节点 %s\n", task.ID, worker.ID)
	return true
}

// CompleteTask 完成任务
func (ts *TaskScheduler) CompleteTask(taskID string, success bool) {
	ts.taskMutex.Lock()
	defer ts.taskMutex.Unlock()

	task, exists := ts.tasks[taskID]
	if !exists {
		return
	}

	now := time.Now()
	task.CompletedAt = &now
	if success {
		task.Status = "completed"
	} else {
		task.Status = "failed"
	}

	// 释放工作节点
	if task.WorkerID != "" {
		ts.workerMutex.Lock()
		if worker, exists := ts.workers[task.WorkerID]; exists {
			worker.Status = "idle"
		}
		ts.workerMutex.Unlock()
	}

	status := "成功"
	if !success {
		status = "失败"
	}
	fmt.Printf("任务 %s 执行%s\n", taskID, status)
}

// Start 启动调度器
func (ts *TaskScheduler) Start() {
	fmt.Println("任务调度器已启动")

	for {
		select {
		case task := <-ts.taskQueue:
			if !ts.Schedule(task) {
				// 重新放回队列
				go func() {
					time.Sleep(1 * time.Second)
					select {
					case ts.taskQueue <- task:
					default:
						fmt.Printf("任务队列仍满: %s\n", task.ID)
					}
				}()
			}
		case <-ts.stopChan:
			fmt.Println("任务调度器已停止")
			return
		}
	}
}

// Stop 停止调度器
func (ts *TaskScheduler) Stop() {
	close(ts.stopChan)
}

// GetTaskStatus 获取任务状态
func (ts *TaskScheduler) GetTaskStatus(taskID string) *Task {
	ts.taskMutex.RLock()
	defer ts.taskMutex.RUnlock()
	return ts.tasks[taskID]
}

// GetClusterStats 获取集群统计信息
func (ts *TaskScheduler) GetClusterStats() map[string]int {
	ts.workerMutex.RLock()
	defer ts.workerMutex.RUnlock()

	stats := make(map[string]int)
	for clusterID := range ts.clusters {
		stats[clusterID] = 0
	}

	for _, worker := range ts.workers {
		if worker.Status == "idle" {
			stats[worker.ClusterID]++
		}
	}

	return stats
}

func main() {
	scheduler := NewTaskScheduler()

	// 添加工作节点
	scheduler.AddWorker(&Worker{ID: "worker1", ClusterID: "cluster1", Status: "idle", Capacity: 1})
	scheduler.AddWorker(&Worker{ID: "worker2", ClusterID: "cluster1", Status: "idle", Capacity: 1})
	scheduler.AddWorker(&Worker{ID: "worker3", ClusterID: "cluster2", Status: "idle", Capacity: 1})

	// 启动调度器
	go scheduler.Start()

	// 提交一些任务
	tasks := []*Task{
		{ID: "task1", Name: "数据处理", ClusterID: "cluster1", Priority: 5},
		{ID: "task2", Name: "文件备份", ClusterID: "cluster2", Priority: 3},
		{ID: "task3", Name: "日志分析", ClusterID: "cluster1", Priority: 7},
		{ID: "task4", Name: "系统监控", ClusterID: "cluster2", Priority: 2},
	}

	for _, task := range tasks {
		scheduler.SubmitTask(task)
	}

	// 模拟任务执行
	time.Sleep(2 * time.Second)

	for _, task := range tasks {
		go func(t *Task) {
			time.Sleep(time.Duration(1+int(time.Now().Unix()%3)) * time.Second)
			scheduler.CompleteTask(t.ID, true)
		}(task)
	}

	// 等待任务完成
	time.Sleep(5 * time.Second)

	// 显示统计信息
	fmt.Println("\n=== 集群统计 ===")
	stats := scheduler.GetClusterStats()
	for cluster, idleWorkers := range stats {
		fmt.Printf("%s: %d 个空闲工作节点\n", cluster, idleWorkers)
	}

	scheduler.Stop()
}
