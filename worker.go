package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	StatusRunning workerStatus = true
	StatusIdle    workerStatus = false
)

type workerStatus bool

type Worker struct {
	id         int
	status     workerStatus
	tasks      <-chan Task
	tCompleted chan<- WorkerResult
	heartbeats chan<- WorkerHeartbeat
	stop       chan struct{}
	stopOnce   sync.Once
}

func CreateWorker(id int, t chan Task, h chan WorkerHeartbeat, _tCompleted chan WorkerResult) *Worker {
	return &Worker{
		id:         id,
		status:     StatusIdle, //TODO: WIP
		tasks:      t,
		tCompleted: _tCompleted,
		heartbeats: h,
		stop:       make(chan struct{}),
	}
}

func (w *Worker) Work() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case t, ok := <-w.tasks:
				if !ok {
					return
				}
				fmt.Printf("Worker %d received message: %s\n", w.id, t.Message)

				// map to tasks assigned
				time.Sleep(time.Second * 10)
				fmt.Printf("Worker %d finished sleeping, sending results\n", w.id)

				w.tCompleted <- WorkerResult{
					worker: w,
					task:   t,
				}
			case <-w.stop:
				return
			}
		}
	}()

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.heartbeats <- WorkerHeartbeat{w.id, time.Now()}
			case <-w.stop:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stop)
	})
}
