package main

import (
	"testing"
	"time"
)

func TestTaskScheduler(t *testing.T) {
	scheduler := NewTaskScheduler()

	// 添加工作节点
	worker1 := &Worker{ID: "worker1", ClusterID: "cluster1", Status: "idle", Capacity: 1}
	worker2 := &Worker{ID: "worker2", ClusterID: "cluster1", Status: "idle", Capacity: 1}
	scheduler.AddWorker(worker1)
	scheduler.AddWorker(worker2)

	// 提交任务
	task := &Task{ID: "task1", Name: "测试任务", ClusterID: "cluster1", Priority: 5}
	scheduler.SubmitTask(task)

	// 启动调度器
	go scheduler.Start()
	defer scheduler.Stop()

	// 等待调度
	time.Sleep(100 * time.Millisecond)

	// 检查任务状态
	scheduledTask := scheduler.GetTaskStatus("task1")
	if scheduledTask.Status != "running" {
		t.Errorf("期望任务状态为running，实际为%s", scheduledTask.Status)
	}

	// 完成任务
	scheduler.CompleteTask("task1", true)

	// 检查任务完成状态
	time.Sleep(100 * time.Millisecond)
	completedTask := scheduler.GetTaskStatus("task1")
	if completedTask.Status != "completed" {
		t.Errorf("期望任务状态为completed，实际为%s", completedTask.Status)
	}

	// 检查工作节点状态
	stats := scheduler.GetClusterStats()
	if stats["cluster1"] != 2 {
		t.Errorf("期望cluster1有2个空闲工作节点，实际有%d个", stats["cluster1"])
	}
}

func TestTaskSubmission(t *testing.T) {
	scheduler := NewTaskScheduler()

	// 提交多个任务
	tasks := []*Task{
		{ID: "task1", Name: "任务1", ClusterID: "cluster1", Priority: 5},
		{ID: "task2", Name: "任务2", ClusterID: "cluster1", Priority: 3},
		{ID: "task3", Name: "任务3", ClusterID: "cluster1", Priority: 1},
	}

	for _, task := range tasks {
		scheduler.SubmitTask(task)
	}

	// 检查任务是否都已提交
	for _, task := range tasks {
		submittedTask := scheduler.GetTaskStatus(task.ID)
		if submittedTask.Status != "pending" {
			t.Errorf("期望任务%s状态为pending，实际为%s", task.ID, submittedTask.Status)
		}
	}
}

func TestWorkerAssignment(t *testing.T) {
	scheduler := NewTaskScheduler()

	// 添加工作节点
	scheduler.AddWorker(&Worker{ID: "worker1", ClusterID: "cluster1", Status: "idle", Capacity: 1})

	// 提交任务到不同集群
	task1 := &Task{ID: "task1", Name: "任务1", ClusterID: "cluster1", Priority: 5}
	task2 := &Task{ID: "task2", Name: "任务2", ClusterID: "cluster2", Priority: 5}

	scheduler.SubmitTask(task1)
	scheduler.SubmitTask(task2)

	// 启动调度器
	go scheduler.Start()
	defer scheduler.Stop()

	time.Sleep(200 * time.Millisecond)

	// cluster1的任务应该能分配到worker1
	scheduledTask1 := scheduler.GetTaskStatus("task1")
	if scheduledTask1.Status != "running" {
		t.Errorf("期望任务1状态为running，实际为%s", scheduledTask1.Status)
	}

	// cluster2的任务应该无法分配（没有worker）
	scheduledTask2 := scheduler.GetTaskStatus("task2")
	if scheduledTask2.Status != "pending" {
		t.Errorf("期望任务2状态为pending，实际为%s", scheduledTask2.Status)
	}
}

func TestClusterStats(t *testing.T) {
	scheduler := NewTaskScheduler()

	// 添加多个集群的工作节点
	scheduler.AddWorker(&Worker{ID: "worker1", ClusterID: "cluster1", Status: "idle", Capacity: 1})
	scheduler.AddWorker(&Worker{ID: "worker2", ClusterID: "cluster1", Status: "busy", Capacity: 1})
	scheduler.AddWorker(&Worker{ID: "worker3", ClusterID: "cluster2", Status: "idle", Capacity: 1})

	stats := scheduler.GetClusterStats()

	if stats["cluster1"] != 1 {
		t.Errorf("期望cluster1有1个空闲工作节点，实际有%d个", stats["cluster1"])
	}

	if stats["cluster2"] != 1 {
		t.Errorf("期望cluster2有1个空闲工作节点，实际有%d个", stats["cluster2"])
	}
}
