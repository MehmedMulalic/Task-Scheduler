package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Coordinator struct {
	logger         *log.Logger
	mux            *http.ServeMux
	mutex          sync.RWMutex
	task           chan<- Task
	wAssigned      <-chan WorkerAssigned
	tCompleted     <-chan WorkerResult
	heartbeats     <-chan WorkerHeartbeat
	wHeartbeats    map[int]time.Time
	wTaskAssigned  map[int]*Task
	taskRepository *TaskRepository
	wCount         int
}

func CreateCoordinator(
	tr *TaskRepository,
	wc int,
	t chan Task,
	h chan WorkerHeartbeat,
	tc chan WorkerResult,
	wa chan WorkerAssigned,
) *Coordinator {
	c := &Coordinator{
		logger:         log.New(os.Stdout, "COORDINATOR: ", log.LstdFlags|log.Lshortfile),
		mux:            http.NewServeMux(),
		taskRepository: tr,
		wCount:         wc,
		task:           t,
		wAssigned:      wa,
		tCompleted:     tc,
		heartbeats:     h,
		wHeartbeats:    map[int]time.Time{},
		wTaskAssigned:  map[int]*Task{},
	}

	c.mux.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	c.mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tasks, err := c.taskRepository.GetAllTasks()
		if err != nil {
			c.logger.Println("Failed to fetch tasks")
			http.Error(w, "failed to fetch tasks", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
	})

	c.mux.HandleFunc("POST /task", c.createTask)

	return c
}

func (c *Coordinator) Run() {
	// Poll pending tasks
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			t, _ := c.taskRepository.GetPendingTask()
			if t != nil {
				c.task <- Task{t.Id, t.Message}
			}
		}
	}()

	// On task completion
	go func() {
		for result := range c.tCompleted {
			c.mutex.Lock()
			delete(c.wTaskAssigned, result.WorkerId)
			c.mutex.Unlock()

			c.taskRepository.UpdateTaskStatus(result.Task.Id, Completed)
			c.logger.Printf("Worker %d completed its task: %s\n", result.WorkerId, result.Task.Message)
		}
	}()

	// On task assigned
	go func() {
		for assignment := range c.wAssigned {
			c.mutex.Lock()
			c.wTaskAssigned[assignment.id] = &assignment.task
			c.mutex.Unlock()
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
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			for wid := range c.wHeartbeats {
				go c.checkWorkerHealth(wid)
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

	err = c.taskRepository.CreateTask(t)
	if err != nil {
		c.logger.Println("Failed to create task")
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("task created\n"))
}

func (c *Coordinator) checkWorkerCount() {
	c.mutex.RLock()
	alive := len(c.wHeartbeats)
	c.mutex.RUnlock()

	if float32(alive) < float32(c.wCount)/2.0 {
		c.logger.Fatal("There are less than half workers alive\n")
	}
}

func (c *Coordinator) checkWorkerHealth(wid int) {
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
			c.taskRepository.UpdateTaskStatus(task.Id, Pending)
			c.logger.Printf("Task reassigned: %s\n", task.Message)
		}

		c.checkWorkerCount()
	}
}
