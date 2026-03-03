package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Task struct {
	Message string `json:"message"`
}

type Coordinator struct {
	mux   *http.ServeMux
	Tasks chan Task
}

func CreateCoordinator() *Coordinator {
	c := &Coordinator{
		mux:   http.NewServeMux(),
		Tasks: make(chan Task, 5),
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

func (c *Coordinator) GetMux() *http.ServeMux {
	return c.mux
}
