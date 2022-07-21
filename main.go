package main

import (
	"flag"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/app"
	"log"
	"net/http"
	"time"
)

//go:generate swagger generate spec -o app/spec/swagger.json

func main() {
	// Command line flags
	port := flag.Uint("port", 8080, "Port to listen on")
	host := flag.String("host", "localhost", "Host to listen on")
	sleep := flag.Int("sleep", 5, "Seconds for a worker to sleep if there is no pending jobs")
	dev := flag.Bool("dev", false, "Run in development mode")
	flag.Parse()

	// Configure the application
	config := app.DefaultConfiguration()
	config.QueueSleepTime = time.Duration(*sleep) * time.Second
	config.Host = *host
	config.Port = *port
	config.DevelopmentMode = *dev

	// Initialize the application
	a, err := app.NewApplication(config)
	if err != nil {
		log.Fatal("error initializing application; ", err)
	}

	defer a.Close()

	// Start the queue processing indefinitely
	go a.ProcessQueue()

	// Start the HTTP server
	addr := a.Addr()
	router := a.GetRouter()
	log.Printf("Server started at %s", addr)
	log.Printf("Development mode: %v", config.DevelopmentMode)
	log.Fatal(http.ListenAndServe(addr, router))
}
