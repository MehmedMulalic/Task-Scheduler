package main

import (
	"fmt"
	"time"
)

const (
	StatusRunning workerStatus = true
	StatusIdle    workerStatus = false
)

type workerStatus bool

type WorkerResult struct {
	worker *Worker
	task   Task
}

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

// TODO: WIP
func (w *Worker) Work() {
	go func() {
		// heartbeats
		w.coordinator.WorkerHeartbeat(w)

		for t := range w.coordinator.Tasks {
			fmt.Printf("Worker %d received message: %s\n", w.id, t.Message)
			// map to tasks assigned
			time.Sleep(time.Second * 10)

			w.coordinator.finished <- WorkerResult{
				worker: w,
				task:   t,
			}
		}
	}()
}
