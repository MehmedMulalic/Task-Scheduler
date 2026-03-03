package main

import "fmt"

type workerStatus bool
type WorkerResult struct {
	worker *Worker
	task   Task
}

const (
	StatusRunning workerStatus = true
	StatusIdle    workerStatus = false
)

type Worker struct {
	coordinator *Coordinator
	id          int
	status      workerStatus
}

func CreateWorker(id int, c *Coordinator) *Worker {
	return &Worker{
		id:          id,
		status:      StatusIdle,
		coordinator: c,
	}
}

func (w *Worker) Work() {
	go func() {
		for t := range w.coordinator.Tasks {
			fmt.Println(t.Message)
			w.coordinator.finished <- WorkerResult{
				worker: w,
				task:   t,
			}
		}
	}()
}
