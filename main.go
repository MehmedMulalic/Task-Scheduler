package main

import (
	"log"
	"net/http"
)

func main() {
	c := CreateCoordinator()
	go c.Run()
	worker := CreateWorker(0, c)
	worker.Work()

	log.Fatal(http.ListenAndServe(":5000", c.GetMux()))
}
