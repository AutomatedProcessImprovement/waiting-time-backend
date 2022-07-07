package app

import (
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"time"
)

type Application struct {
	router *mux.Router

	queue *Queue
}

func NewApplication() *Application {
	app := &Application{}
	app.queue = NewQueue()
	app.initializeRouter()
	return app
}

func (app *Application) Router() *mux.Router {
	return app.router
}

func (app *Application) processJob(job *Job) {
	job.Status = JobStatusRunning

	jobDuration := time.Duration(rand.Intn(60)+60) * time.Second
	time.Sleep(jobDuration)

	job.Status = JobStatusCompleted

	now := time.Now()
	job.CompletedAt = &now

	log.Printf("Job %s completed", job.ID)
}
