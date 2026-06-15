package core

import (
	"log"
	"sync"
)

type TaskType string

const (
	TaskScan   TaskType = "Library Scan"
	TaskScrape TaskType = "Auto Scrape"
)

type Task struct {
	ID       string   `json:"id"`
	Type     TaskType `json:"type"`
	Status   string   `json:"status"` // pending, running, completed, failed
	Message  string   `json:"message"`
	Progress int      `json:"progress"`
	Total    int      `json:"total"`
	Work     func()   `json:"-"`
}

type TaskManager struct {
	tasks      []*Task
	taskChan   chan *Task
	mu         sync.RWMutex
}

var GlobalTaskManager *TaskManager

func InitTaskManager() {
	GlobalTaskManager = &TaskManager{
		tasks:    make([]*Task, 0),
		taskChan: make(chan *Task, 100),
	}
	go GlobalTaskManager.worker()
}

func (tm *TaskManager) AddTask(t *Task) {
	tm.mu.Lock()
	t.Status = "pending"
	tm.tasks = append(tm.tasks, t)
	tm.mu.Unlock()
	tm.taskChan <- t
}

func (tm *TaskManager) worker() {
	for t := range tm.taskChan {
		tm.updateTask(t, "running", "")
		log.Printf("Starting task: %s (%s)", t.Type, t.ID)
		t.Work()
		tm.updateTask(t, "completed", "Done")
		log.Printf("Finished task: %s (%s)", t.Type, t.ID)
	}
}

func (tm *TaskManager) updateTask(t *Task, status, msg string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	t.Status = status
	t.Message = msg
}

func (tm *TaskManager) UpdateProgress(t *Task, progress, total int, msg string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	t.Progress = progress
	t.Total = total
	if msg != "" {
		t.Message = msg
	}
}

func (tm *TaskManager) GetTasks() []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	// Return last 10 tasks
	start := 0
	if len(tm.tasks) > 10 {
		start = len(tm.tasks) - 10
	}
	return tm.tasks[start:]
}
