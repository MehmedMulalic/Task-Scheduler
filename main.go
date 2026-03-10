package main

import (
	"log"
	"net/http"
)

func main() {
	WORKER_COUNT := 3

	tasks := make(chan Task, 5)
	heartbeats := make(chan WorkerHeartbeat)
	tCompleted := make(chan WorkerResult)
	assignedWorker := make(chan WorkerAssigned)

	c := CreateCoordinator(WORKER_COUNT, tasks, heartbeats, tCompleted, assignedWorker)
	c.Run()

	workers := make([]*Worker, 0, WORKER_COUNT)
	for i := 0; i < WORKER_COUNT; i++ {
		workers = append(workers, CreateWorker(i, tasks, heartbeats, tCompleted, assignedWorker))
		workers[i].Work()
	}

	log.Println("Server initiated at port 5000")
	log.Fatal(http.ListenAndServe(":5000", c.GetMux()))
}
