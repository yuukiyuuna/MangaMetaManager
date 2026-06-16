package core

import "testing"

func TestAddTaskIfNoActiveTypeRejectsActiveTask(t *testing.T) {
	tm := &TaskManager{
		tasks:    make([]*Task, 0),
		taskChan: make(chan *Task, 2),
	}

	first := &Task{ID: "scan-1", Type: TaskScan}
	if !tm.AddTaskIfNoActiveType(first) {
		t.Fatal("expected first scan task to be accepted")
	}

	second := &Task{ID: "scan-2", Type: TaskScan}
	if tm.AddTaskIfNoActiveType(second) {
		t.Fatal("expected duplicate active scan task to be rejected")
	}
}

func TestAddTaskIfNoActiveTypeAllowsCompletedTaskType(t *testing.T) {
	tm := &TaskManager{
		tasks: []*Task{
			{ID: "scan-1", Type: TaskScan, Status: "completed"},
		},
		taskChan: make(chan *Task, 1),
	}

	if !tm.AddTaskIfNoActiveType(&Task{ID: "scan-2", Type: TaskScan}) {
		t.Fatal("expected scan task to be accepted after previous task completed")
	}
}
