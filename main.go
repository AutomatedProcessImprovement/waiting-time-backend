package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/AutomatedProcessImprovement/waiting-time-backend/app"
)

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	log.Printf("Server started")

	a := app.NewApplication()
	router := a.Router()

	addr := fmt.Sprintf(":%s", *port)
	log.Fatal(http.ListenAndServe(addr, router))
}
