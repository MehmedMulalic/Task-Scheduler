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
	wAssigned     <-chan WorkerAssigned
	tCompleted    <-chan WorkerResult
	heartbeats    <-chan WorkerHeartbeat
	wHeartbeats   map[int]time.Time
	wTaskAssigned map[int]*Task
	wCount        int
}

func CreateCoordinator(wc int, t chan Task, h chan WorkerHeartbeat, tc chan WorkerResult, wa chan WorkerAssigned) *Coordinator {
	c := &Coordinator{
		logger:        log.New(os.Stdout, "COORDINATOR: ", log.LstdFlags|log.Lshortfile),
		mux:           http.NewServeMux(),
		task:          t,
		wAssigned:     wa,
		tCompleted:    tc,
		heartbeats:    h,
		wHeartbeats:   map[int]time.Time{},
		wTaskAssigned: map[int]*Task{},
		wCount:        wc,
	}

	c.mux.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Healthy")
	})

	c.mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "WIP - I got tasks")
	})

	c.mux.HandleFunc("POST /task", c.createTask)

	return c
}

func (c *Coordinator) Run() {
	ticker := time.NewTicker(5 * time.Second)

	// On task completion
	go func() {
		for result := range c.tCompleted {
			c.mutex.Lock()
			delete(c.wTaskAssigned, result.id)
			c.mutex.Unlock()

			c.logger.Printf("Worker %d completed its task: %s\n", result.id, result.task.Message)
		}
	}()

	// On task assigned
	go func() {
		for assignment := range c.wAssigned {
			c.mutex.Lock()
			c.wTaskAssigned[assignment.id] = &assignment.task
			c.mutex.Unlock()

			fmt.Println(c.wTaskAssigned[assignment.id]) //TODO: test, delete
		}
	}()

	// On worker hearbeat
	go func() {
		for heartbeat := range c.heartbeats {
			c.mutex.Lock()
			c.wHeartbeats[heartbeat.workerId] = heartbeat.time
			c.mutex.Unlock()

			c.logger.Printf("Worker ID %d pingged\n", heartbeat.workerId)
		}
	}()

	// On worker death
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
						task := c.wTaskAssigned[wid]
						delete(c.wTaskAssigned, wid)
						c.mutex.Unlock()

						if task != nil {
							select {
							case c.task <- *task:
								c.logger.Printf("Task reassigned: %s\n", task.Message)
							default:
								c.logger.Printf("Task queue full, could not reassign: %s\n", task.Message)
							}
						}

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
	c.mutex.RLock()
	alive := len(c.wHeartbeats)
	c.mutex.RUnlock()

	if float32(alive) < float32(c.wCount)/2.0 {
		c.logger.Fatal("There are less than half workers alive\n")
	}
}
