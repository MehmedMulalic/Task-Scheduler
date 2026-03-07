package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Coordinator struct {
	mux           *http.ServeMux
	Tasks         chan Task
	finished      chan WorkerResult
	mut           sync.Mutex
	wHeartbeats   map[int]time.Time
	wTaskAssigned map[int]*Task
}

type Task struct {
	Message string `json:"message"`
}

func CreateCoordinator() *Coordinator {
	c := &Coordinator{
		mux:           http.NewServeMux(),
		Tasks:         make(chan Task, 5),
		finished:      make(chan WorkerResult),
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
	for result := range c.finished {
		fmt.Printf("Worker %d completed task: %s\n", result.worker.id, result.task.Message)
	}
}

func (c *Coordinator) GetMux() *http.ServeMux {
	return c.mux
}

// TODO: WIP
func (c *Coordinator) WorkerHeartbeat(w *Worker) {
	c.wHeartbeats[w.id] = time.Now()

}

func (c *Coordinator) createTask(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type must be json", http.StatusUnsupportedMediaType)
		return
	}

	var task Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received: %+v\n", task)
	w.WriteHeader(http.StatusCreated)

	c.Tasks <- task
}

// TODO: WIP
func (c *Coordinator) initiateHeartbeat(w *Worker) {
	go func() {
		time.Sleep(5 * time.Second)

	}()
}
