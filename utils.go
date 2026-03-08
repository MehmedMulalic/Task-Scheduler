package main

import "time"

type Task struct {
	Message string `json:"message"`
}

type WorkerHeartbeat struct {
	workerId int
	time     time.Time
}

type WorkerResult struct {
	worker *Worker
	task   Task
}
