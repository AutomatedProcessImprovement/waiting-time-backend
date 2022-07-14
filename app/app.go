package app

import (
	"context"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/gorilla/mux"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type Application struct {
	router *mux.Router
	queue  *Queue
	config *Configuration
	logger *log.Logger
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

	app.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	if err := mkdir(config.ResultsDir); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *Application) Addr() string {
	return fmt.Sprintf("0.0.0.0:%d", app.config.Port)
}

func (app *Application) Close() {
}

func (app *Application) Router() *mux.Router {
	return app.router
}

func (app *Application) AddJob(job *model.Job) error {
	return app.queue.Add(job)
}

func (app *Application) ProcessQueue() {
	app.logger.Printf("Queue processing started")

	for {
		job := app.queue.Next()
		if job == nil {
			time.Sleep(app.config.QueueSleepTime)
			continue
		}

		app.processJob(job)

		if err := app.SaveQueue(); err != nil {
			app.logger.Printf("error saving queue: %s", err.Error())
		}
	}
}

func (app *Application) SaveQueue() error {
	app.queue.lock.Lock()
	defer app.queue.lock.Unlock()
	return dumpGob(app.config.QueuePath, app.queue, app.logger)
}

func (app *Application) LoadQueue() error {
	app.queue.lock.Lock()
	defer app.queue.lock.Unlock()
	return readGob(app.config.QueuePath, app.queue, app.logger)
}

func (app *Application) processJob(job *model.Job) {
	// check for a pending job
	if job.Status != model.JobStatusPending {
		err := fmt.Errorf("job is not pending")
		app.logger.Printf("Job %s failed; %s", job.ID, err.Error())
		job.SetError(err)
		return
	}

	// pre-work
	var eventLogName string
	{
		app.logger.Printf("Job %s started", job.ID)
		job.SetStatus(model.JobStatusRunning)

		// job's directory
		if err := mkdir(job.Dir); err != nil {
			app.logger.Printf("error creating job's directory: %s", err.Error())
			job.SetError(err)
			return
		}

		// download log into job.Dir
		eventLogURL := job.EventLog.String()
		eventLogName = path.Base(eventLogURL)
		eventLogPath := path.Join(job.Dir, eventLogName)
		if err := download(eventLogURL, eventLogPath, app.logger); err != nil {
			app.logger.Printf("error downloading event log: %s", err.Error())
			job.SetError(err)
			return
		}
	}

	// work
	{
		ctx, cancel := context.WithTimeout(context.Background(), app.config.JobTimeout)
		defer cancel()

		jobErrorChan := make(chan error)
		go func() {
			jobErrorChan <- app.runAnalysis(ctx, eventLogName, job.Dir, job.ID)
		}()

		select {
		case <-ctx.Done():
			app.logger.Printf("Job %s timed out", job.ID)
			job.SetError(fmt.Errorf("job timed out"))
			job.SetStatus(model.JobStatusFailed)
		case jobError := <-jobErrorChan:
			if jobError != nil {
				app.logger.Printf("Job %s failed; %s", job.ID, jobError.Error())
				job.SetError(jobError)
				job.SetStatus(model.JobStatusFailed)
			} else {
				app.logger.Printf("Job %s completed", job.ID)
				job.SetStatus(model.JobStatusCompleted)

				// assign report CSV
				ext := path.Ext(eventLogName)
				reportName := strings.TrimSuffix(eventLogName, ext) + "_handoff" + ext
				reportURL, err := url.Parse(
					fmt.Sprintf("http://%s:%d/assets/results/%s/%s",
						app.config.Host, app.config.Port, job.ID, reportName))
				if err != nil {
					app.logger.Printf("error creating report URL: %s", err.Error())
					job.SetError(err)
				}
				job.SetReportCSV(&model.URL{URL: reportURL})
			}
		}
	}

	// post-work
	job.SetCompletedAt(time.Now())

	// TODO: compose JobResult
}

func (app *Application) runAnalysis(ctx context.Context, eventLogName, jobDir, jobID string) error {
	jobDir, err := abspath(jobDir)
	if err != nil {
		return err
	}

	eventLogPath := path.Join(jobDir, eventLogName)
	scriptName := "run_analysis.bash"
	if app.config.DevelopmentMode {
		scriptName = "run_analysis_dev.bash"
	}
	args := fmt.Sprintf("bash %s %s %s", scriptName, eventLogPath, jobDir)

	out, err := exec.CommandContext(ctx, "sh", "-c", args).Output()
	if len(out) > 0 {
		app.logger.Printf("Job %s output: %s", jobID, out)
	}

	return err
}
