package main

import (
	"log"
	"os"
	"sync"
	"time"
)

const (
	StatusRunning workerStatus = true
	StatusIdle    workerStatus = false
)

type workerStatus bool

type Worker struct {
	logger     *log.Logger
	id         int
	status     workerStatus
	tasks      <-chan Task
	wAssigned  chan<- WorkerAssigned
	heartbeats chan<- WorkerHeartbeat
	tCompleted chan<- WorkerResult
	stop       chan struct{}
	stopOnce   sync.Once
}

func CreateWorker(id int, t chan Task, h chan WorkerHeartbeat, tc chan WorkerResult, wa chan WorkerAssigned) *Worker {
	return &Worker{
		logger:     log.New(os.Stdout, "WORKER: ", log.LstdFlags|log.Lshortfile),
		id:         id,
		status:     StatusIdle, //TODO: WIP
		wAssigned:  wa,
		tCompleted: tc,
		tasks:      t,
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
				w.logger.Printf("Worker %d received message: %s\n", w.id, t.Message)
				w.wAssigned <- WorkerAssigned{w.id, t}

				time.Sleep(time.Second * 10)
				w.logger.Printf("Worker %d finished sleeping, sending results\n", w.id)

				w.tCompleted <- WorkerResult{
					id:   w.id,
					task: t,
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
