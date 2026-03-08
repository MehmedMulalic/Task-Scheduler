package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	tasks := make(chan Task, 5)
	heartbeats := make(chan WorkerHeartbeat)
	tCompleted := make(chan WorkerResult)

	c := CreateCoordinator(tasks, heartbeats, tCompleted)
	c.Run()

	workers := make([]*Worker, 0, 3)
	for i := 0; i < 3; i++ {
		workers = append(workers, CreateWorker(i, tasks, heartbeats, tCompleted))
		workers[i].Work()
	}

	// Kill worker test
	time.Sleep(7 * time.Second)
	fmt.Println("Stopping worker. Time:", time.Now())
	workers[0].Stop()

	log.Fatal(http.ListenAndServe(":5000", c.GetMux()))
}
