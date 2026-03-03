package main

import "fmt"

type workerStatus bool

const (
	StatusRunning workerStatus = true
	StatusIdle    workerStatus = false
)

type event string

type Worker struct {
	coordinator *Coordinator
	id          int
	status      workerStatus
	channel     chan event
}

func CreateWorker(id int, c *Coordinator) *Worker {
	return &Worker{
		id:          id,
		status:      StatusIdle,
		channel:     make(chan event),
		coordinator: c,
	}
}

func (w *Worker) Work() {
	go func() {
		for task := range w.coordinator.Tasks {
			fmt.Println(task.Message)
		}
	}()
}

func (w *Worker) GetChannel() chan event {
	return w.channel
}
