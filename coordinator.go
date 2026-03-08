package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	logger        *log.Logger
	mux           *http.ServeMux
	mutex         sync.RWMutex
	task          chan<- Task
	tCompleted    <-chan WorkerResult
	heartbeats    <-chan WorkerHeartbeat
	wHeartbeats   map[int]time.Time
	wTaskAssigned map[int]*Task
}

func CreateCoordinator(t chan Task, h chan WorkerHeartbeat, _tCompleted chan WorkerResult) *Coordinator {
	c := &Coordinator{
		logger:        log.New(os.Stdout, "COORDINATOR: ", log.LstdFlags|log.Lshortfile),
		mux:           http.NewServeMux(),
		task:          t,
		tCompleted:    _tCompleted,
		heartbeats:    h,
		wHeartbeats:   map[int]time.Time{},
		wTaskAssigned: map[int]*Task{},
	}

	c.mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This is a test")
	})

	c.mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "I got tasks")
	})

	c.mux.HandleFunc("POST /task", c.createTask)

	return c
}

func (c *Coordinator) Run() {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for result := range c.tCompleted {
			c.logger.Printf("Worker %d completed its task: %s\n", result.worker.id, result.task.Message)
		}
	}()

	go func() {
		for heartbeat := range c.heartbeats {
			c.mutex.Lock()
			c.wHeartbeats[heartbeat.workerId] = heartbeat.time
			c.mutex.Unlock()

			c.logger.Printf("Worker ID %d pingged\n", heartbeat.workerId)
		}
	}()

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			for wid := range c.wHeartbeats {
				go func(wid int) {
					c.mutex.RLock()
					workerHeartbeat := c.wHeartbeats[wid]
					c.mutex.RUnlock()

					if time.Since(workerHeartbeat) > 10*time.Second {
						c.logger.Printf("Worker %d died.\n", wid)
						c.mutex.Lock()
						delete(c.wHeartbeats, wid)
						c.mutex.Unlock()

						c.checkWorkerCount()
					}
				}(wid)
			}
		}
	}()
}

func (c *Coordinator) GetMux() *http.ServeMux {
	return c.mux
}

func (c *Coordinator) createTask(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type must be json", http.StatusUnsupportedMediaType)
		return
	}

	var t Task

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	c.logger.Printf("Received: %+v\n", t)
	w.WriteHeader(http.StatusCreated)

	c.task <- t
}

func (c *Coordinator) checkWorkerCount() {
	//TODO: WIP
}
