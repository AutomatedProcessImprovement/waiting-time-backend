// Package app Waiting Time Analysis Backend API
//
// The tool allows to identify activity transitions given an event log and analyze its waiting times.
//
// Schemes: http
// Host: 193.40.11.233
// BasePath: /
// Version: 1.0.0
//
// Consumes:
//     - application/json
//
// Produces:
// 		- application/json
//
// swagger:meta
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

func (app *Application) GetRouter() *mux.Router {
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
		eventLogURL := job.EventLogURL.String()
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

		host := os.Getenv("WEBAPP_HOST")
		if len(host) == 0 {
			host = app.config.Host
		}

		jobErrorChan := make(chan error)
		go func() {
			jobErrorChan <- app.runAnalysis(ctx, eventLogName, job.Dir, job.ID)
		}()

		const reportSuffixCSV = "_transitions_report.csv"

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
				reportName := strings.TrimSuffix(eventLogName, ext) + reportSuffixCSV
				reportURL, err := url.Parse(
					fmt.Sprintf("http://%s/assets/results/%s/%s",
						host, job.ID, reportName))
				if err != nil {
					app.logger.Printf("error creating report URL: %s", err.Error())
					job.SetError(err)
				}
				job.SetReportCSV(&model.URL{URL: reportURL})

				// assign result
				result, err := app.prepareJobResult(job)
				if err != nil {
					app.logger.Printf("error preparing result: %s", err.Error())
					job.SetError(err)
				} else {
					job.SetResult(result)
				}
			}
		}
	}

	// post-work
	job.SetCompletedAt(time.Now())
}

func (app *Application) prepareJobResult(job *model.Job) (*model.JobResult, error) {
	if job.EventLogURL == nil {
		return nil, fmt.Errorf("job has no event log")
	}

	const (
		reportSuffixJSON = "_transitions_report.json"
		cteSuffix        = "_process_cte_impact.json"
	)

	eventLogName := path.Base(job.EventLogURL.String())
	eventLogExt := path.Ext(eventLogName)

	// prepare result
	resultName := strings.TrimSuffix(eventLogName, eventLogExt) + reportSuffixJSON
	resultPath := path.Join(job.Dir, resultName)
	result := model.JobResult{}
	if err := readJSON(resultPath, &result, app.logger); err != nil {
		return nil, fmt.Errorf("error reading result: %s", err.Error())
	}

	// assign CTE impact
	cteName := strings.TrimSuffix(eventLogName, eventLogExt) + cteSuffix
	ctePath := path.Join(job.Dir, cteName)
	cteImpact := model.JobCteImpact{}
	if err := readJSON(ctePath, &cteImpact, app.logger); err != nil {
		return nil, fmt.Errorf("error reading CTE impact: %s", err.Error())
	} else {
		result.CTEImpact = &cteImpact
	}

	return &result, nil
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
