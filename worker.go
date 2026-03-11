package main

import (
	"log"
	"os"
	"sync"
	"time"
)

type Worker struct {
	logger     *log.Logger
	id         int
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
		wAssigned:  wa,
		tCompleted: tc,
		tasks:      t,
		heartbeats: h,
		stop:       make(chan struct{}),
	}
}

func (w *Worker) Work() {
	go func() {
		for {
			select {
			case t, ok := <-w.tasks:
				if !ok {
					w.logger.Println("Tasks channel closed")
					return
				}
				w.logger.Printf("Worker %d received message: %s\n", w.id, t.Message)
				w.wAssigned <- WorkerAssigned{w.id, t}

				select {
				case <-time.After(time.Second * 15):
					w.logger.Printf("Worker %d finished sleeping, sending results\n", w.id)
					w.tCompleted <- WorkerResult{
						w.id,
						t,
					}
				case <-w.stop:
					w.logger.Printf("Worker %d stopped\n", w.id)
					return
				}

			case <-w.stop:
				w.logger.Printf("Worker %d stopped\n", w.id)
				return
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
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
