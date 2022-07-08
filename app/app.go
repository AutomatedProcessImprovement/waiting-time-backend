package app

import (
	"encoding/gob"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/gorilla/mux"
	"log"
	"os"
	"time"
)

type Application struct {
	router    *mux.Router
	queue     *Queue
	config    *Configuration
	appLogger *log.Logger
	webLogger *log.Logger

	appLogFile     *os.File
	requestLogFile *os.File
}

func NewApplication(config *Configuration) (*Application, error) {
	app := &Application{
		config: config,
		queue:  NewQueue(),
	}

	err := app.LoadQueue()
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("error loading queue: %s", err.Error())
	}

	app.initializeRouter()

	appFile, err := os.OpenFile(config.AppLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	app.appLogFile = appFile
	app.appLogger = log.New(appFile, "", log.Ldate|log.Ltime)

	requestsFile, err := os.OpenFile(config.RequestLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	app.requestLogFile = requestsFile
	app.webLogger = log.New(requestsFile, "", log.Ldate|log.Ltime)

	return app, nil
}

func (app *Application) Close() {
	if err := app.appLogFile.Close(); err != nil {
		app.appLogger.Printf("error closing file: %s", err.Error())
	}
}

func (app *Application) Router() *mux.Router {
	return app.router
}

func (app *Application) AddJob(job *model.Job) error {
	return app.queue.Add(job)
}

func (app *Application) ProcessQueue() {
	app.appLogger.Printf("Queue processing started")

	for {
		job := app.queue.Next()
		if job == nil {
			time.Sleep(app.config.WorkerSleepTime)
			continue
		}
		job.Run(app.appLogger)

		if err := app.SaveQueue(); err != nil {
			app.appLogger.Printf("error saving queue: %s", err.Error())
		}
	}
}

func (app *Application) SaveQueue() error {
	app.queue.lock.Lock()
	defer app.queue.lock.Unlock()
	return dumpGob(app.config.QueueStorePath, app.queue, app.appLogger)
}

func (app *Application) LoadQueue() error {
	app.queue.lock.Lock()
	defer app.queue.lock.Unlock()
	return readGob(app.config.QueueStorePath, app.queue, app.appLogger)
}

func dumpGob(path string, data interface{}, logger *log.Logger) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Printf("error closing file: %s", err.Error())
		}
	}()

	return gob.NewEncoder(f).Encode(data)
}

func readGob(path string, data interface{}, logger *log.Logger) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Printf("error closing file: %s", err.Error())
		}
	}()

	return gob.NewDecoder(f).Decode(data)
}
