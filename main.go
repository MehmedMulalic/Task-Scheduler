package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

const WORKER_COUNT = 3

func main() {
	logger := log.New(os.Stdout, "MAIN: ", log.LstdFlags|log.Lshortfile)

	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		logger.Fatal("Database could not be opened. Error message: ", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending'
		)
	`)
	if err != nil {
		logger.Fatal(err)
	}

	taskRepository := NewTaskRepository(db)
	tasks := make(chan Task, 5)
	heartbeats := make(chan WorkerHeartbeat)
	tCompleted := make(chan WorkerResult)
	assignedWorker := make(chan WorkerAssigned)

	c := CreateCoordinator(taskRepository, WORKER_COUNT, tasks, heartbeats, tCompleted, assignedWorker)
	c.Run()

	workers := make([]*Worker, 0, WORKER_COUNT)
	for i := 0; i < WORKER_COUNT; i++ {
		workers = append(workers, CreateWorker(i, tasks, heartbeats, tCompleted, assignedWorker))
		workers[i].Work()
	}

	logger.Println("Server initiated at port 5000")
	logger.Fatal(http.ListenAndServe(":5000", c.GetMux()))
}
