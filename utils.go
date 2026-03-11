package main

import "time"

const (
	Pending    TaskStatus = "pending"
	InProgress TaskStatus = "in_progress"
	Completed  TaskStatus = "completed"
)

type TaskStatus string

type Task struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

type TaskResponse struct {
	Id      int        `json:"task_id"`
	Message string     `json:"task_message"`
	Status  TaskStatus `json:"status"`
}

type WorkerAssigned struct {
	id   int
	task Task
}

type WorkerHeartbeat struct {
	workerId int
	time     time.Time
}

type WorkerResult struct {
	WorkerId int
	Task     Task
}
