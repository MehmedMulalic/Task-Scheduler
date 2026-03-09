package main

import "time"

type Task struct {
	Message string `json:"message"`
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
	id   int
	task Task
}
