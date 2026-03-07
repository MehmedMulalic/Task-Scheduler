package main

import (
	"log"
	"net/http"
)

func main() {
	heartbeats := make(chan workerHeartbeat)

	c := CreateCoordinator(heartbeats)
	c.Run()

	workers := make([]*Worker, 0, 3)
	for i := range 3 {
		workers = append(workers, CreateWorker(i, c, heartbeats))
		workers[i].Work()
	}

	log.Fatal(http.ListenAndServe(":5000", c.GetMux()))
}
