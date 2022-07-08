package main

import (
	"flag"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/app"
	"log"
	"net/http"
	"time"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to listen on")
	sleep := flag.Int("sleep", 5, "Seconds for a worker to sleep if there is no pending jobs")
	flag.Parse()

	// Configure the application
	config := app.DefaultConfiguration()
	config.QueueSleepTime = time.Duration(*sleep) * time.Second

	// Initialize the application
	a, err := app.NewApplication(config)
	if err != nil {
		log.Fatal("error initializing application; ", err)
	}

	defer a.Close()

	// Start the queue processing indefinitely
	go a.ProcessQueue()

	// Start the HTTP server
	addr := fmt.Sprintf(":%s", *port)
	router := a.Router()
	log.Printf("Server started at %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
