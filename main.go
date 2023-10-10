package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/app"
	"log"
	"net/http"
	"os"
	"time"
)

//go:generate swagger generate spec -o app/spec/swagger.json -m

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
	// Database connection check.
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open a DB connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	log.Fatal(http.ListenAndServe(addr, router))
}
