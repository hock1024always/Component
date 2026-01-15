# 多集群任务调度器

这是一个简单的多集群任务调度器，支持跨集群的任务分配和负载均衡。

## 架构设计

### 核心组件

1. **Task** - 任务结构体
   - ID: 任务唯一标识
   - Name: 任务名称
   - ClusterID: 所属集群
   - Status: 任务状态 (pending, running, completed, failed)
   - Priority: 优先级 (1-10)
   - WorkerID: 执行该任务的工作节点

2. **Worker** - 工作节点结构体
   - ID: 工作节点唯一标识
   - ClusterID: 所属集群
   - Status: 状态 (idle, busy)
   - Capacity: 并发处理能力

3. **TaskScheduler** - 任务调度器
   - tasks: 任务存储
   - workers: 工作节点存储
   - clusters: 集群到工作节点的映射
   - taskQueue: 任务队列

## 调度策略

1. **集群优先**: 优先在本集群内分配任务
2. **负载均衡**: 在集群内部进行负载均衡
3. **跨集群调度**: 本集群无可用节点时，在其他集群寻找
4. **优先级支持**: 支持任务优先级调度

## 代码结构解析

### Task 结构体详解

```go
type Task struct {
    ID          string        // 任务ID
    Name        string        // 任务名称
    ClusterID   string        // 目标集群
    Status      string        // 任务状态
    Priority    int           // 优先级 1-10
    CreatedAt   time.Time     // 创建时间
    StartedAt   *time.Time    // 开始时间
    CompletedAt *time.Time    // 完成时间
    WorkerID    string        // 执行节点ID
}
```

### Worker 结构体详解

```go
type Worker struct {
    ID        string // 节点ID
    ClusterID string // 集群ID
    Status    string // 状态: idle/busy
    Capacity  int    // 并发能力
}
```

### TaskScheduler 核心调度逻辑

**SubmitTask 方法**: 任务提交
```go
func (ts *TaskScheduler) SubmitTask(task *Task) {
    ts.taskMutex.Lock()
    task.Status = "pending"        // 设置初始状态
    task.CreatedAt = time.Now()    // 记录创建时间
    ts.tasks[task.ID] = task       // 存储任务
    ts.taskMutex.Unlock()

    select {
    case ts.taskQueue <- task:     // 尝试加入队列
        fmt.Printf("任务已提交: %s\n", task.ID)
    default:                       // 队列已满
        fmt.Printf("任务队列已满，稍后重试: %s\n", task.ID)
    }
}
```

**Schedule 方法**: 任务调度核心逻辑
```go
func (ts *TaskScheduler) Schedule(task *Task) bool {
    // 1. 优先在本集群寻找空闲节点
    clusterWorkers := ts.clusters[task.ClusterID]
    for _, workerID := range clusterWorkers {
        worker := ts.workers[workerID]
        if worker.Status == "idle" {
            return ts.assignTask(task, worker)
        }
    }

    // 2. 在其他集群寻找空闲节点
    for clusterID, workerIDs := range ts.clusters {
        if clusterID == task.ClusterID {
            continue // 跳过本集群
        }
        for _, workerID := range workerIDs {
            worker := ts.workers[workerID]
            if worker.Status == "idle" {
                return ts.assignTask(task, worker)
            }
        }
    }

    return false // 没有找到合适的节点
}
```

**assignTask 方法**: 任务分配
```go
func (ts *TaskScheduler) assignTask(task *Task, worker *Worker) bool {
    // 1. 双重检查worker状态 (防止并发竞争)
    if worker.Status != "idle" {
        return false
    }

    // 2. 分配任务
    worker.Status = "busy"
    now := time.Now()
    task.Status = "running"
    task.StartedAt = &now
    task.WorkerID = worker.ID

    return true
}
```

**Start 方法**: 调度器主循环
```go
func (ts *TaskScheduler) Start() {
    for {
        select {
        case task := <-ts.taskQueue:     // 从队列获取任务
            if !ts.Schedule(task) {     // 尝试调度
                // 调度失败，1秒后重试
                go func() {
                    time.Sleep(1 * time.Second)
                    ts.taskQueue <- task
                }()
            }
        case <-ts.stopChan:            // 接收停止信号
            return
        }
    }
}
```

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

### 2. 运行测试
```bash
go test -v
```

## 测试覆盖

- `TestTaskScheduler`: 测试基本调度功能
- `TestTaskSubmission`: 测试任务提交
- `TestWorkerAssignment`: 测试工作节点分配
- `TestClusterStats`: 测试集群统计

## 扩展思路

1. **持久化**: 将任务状态持久化到数据库
2. **健康检查**: 添加工作节点健康检查机制
3. **动态扩缩容**: 支持动态添加/移除工作节点
4. **任务依赖**: 支持任务间的依赖关系
5. **监控告警**: 添加调度指标监控和告警