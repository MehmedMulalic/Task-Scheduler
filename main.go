package main

import (
	"log"
	"net/http"
)

func main() {
	c := CreateCoordinator()
	go c.Run()

	workers := make([]Worker, 0, 3)
	for i := range 3 {
		workers = append(workers, *CreateWorker(i, c))
		workers[i].Work()
	}

	log.Fatal(http.ListenAndServe(":5000", c.GetMux()))
}
