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

type workerHeartbeat struct {
	workerId int
	time     time.Time
}

type WorkerResult struct {
	worker *Worker
	task   Task
}

type Worker struct {
	coordinator *Coordinator
	id          int
	status      workerStatus
	heartbeat   chan<- workerHeartbeat
}

func CreateWorker(id int, c *Coordinator, heartbeats chan workerHeartbeat) *Worker {
	return &Worker{
		id:          id,
		status:      StatusIdle,
		coordinator: c, //TODO: treba li coordinator u worker?
		heartbeat:   heartbeats,
	}
}

func (w *Worker) Work() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
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

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			w.heartbeat <- workerHeartbeat{w.id, time.Now()}
		}
	}()
}
