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
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type Application struct {
	router *mux.Router
	queue  *Queue
	config *Configuration
	logger *log.Logger

	analysisCancelFunc context.CancelFunc
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
	err := readGob(app.config.QueuePath, app.queue, app.logger)
	if os.IsNotExist(err) {
		err = nil
	} else if err != nil {
		return fmt.Errorf("error loading queue: %s", err.Error())
	}
	if app.queue.Jobs == nil {
		app.queue.Jobs = []*model.Job{}
	}
	return err
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
	var eventLogName = path.Base(job.EventLogURL.String())
	{
		app.logger.Printf("Job %s started", job.ID)
		job.SetStatus(model.JobStatusRunning)

		eventLogPath := path.Join(job.Dir, eventLogName)

		// if the job was created from a request body, then the even log file is already downloaded
		if !job.EventLogFromRequestBody {
			// job's directory
			if err := mkdir(job.Dir); err != nil {
				app.logger.Printf("error creating job's directory: %s", err.Error())
				job.SetError(err)
				return
			}

			// download log into job.Dir
			if err := download(job.EventLogURL.String(), eventLogPath, app.logger); err != nil {
				app.logger.Printf("error downloading event log: %s", err.Error())
				job.SetError(err)
				return
			}
		}

		// make MD5 hash of the log to check for uniqueness of the file
		job.EventLogMD5, _ = md5sum(eventLogPath) // NOTE: we can ignore the error here

		// if the log has been processed before, skip analysis and assign the result to the job
		foundJob := app.queue.FindByMD5(job.EventLogMD5)
		if foundJob != nil && foundJob.ID != job.ID && foundJob.Status != model.JobStatusPending {
			app.logger.Printf("Job %s skipped; log has been processed before", job.ID)
			job.SetStatus(model.JobStatusDuplicate)
			job.SetResult(foundJob.Result)
			job.SetReportCSV(foundJob.ReportCSV)
			job.SetCompletedAt(time.Now())
			if foundJob.Error != "" {
				job.SetError(errors.New(foundJob.Error))
			}
			return
		}
	}

	// work
	{
		ctx, cancel := context.WithTimeout(context.Background(), app.config.JobTimeout)
		defer cancel()
		// gives control over the running analysis process to the whole app
		app.analysisCancelFunc = cancel

		host := os.Getenv("WEBAPP_HOST")
		if len(host) == 0 {
			host = app.config.Host
		}

		jobErrorChan := make(chan error)
		go func() {
			jobErrorChan <- app.runAnalysis(ctx, eventLogName, job)
		}()

		const reportSuffixCSV = "_transitions_report.csv"

		select {
		case <-ctx.Done():
			app.logger.Printf("Job %s has been interrupted", job.ID)
			job.SetError(fmt.Errorf("job has been interrupted"))
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

func (app *Application) runAnalysis(ctx context.Context, eventLogName string, job *model.Job) error {
	jobDir, err := abspath(job.Dir)
	if err != nil {
		return err
	}

	eventLogPath := path.Join(jobDir, eventLogName)
	scriptName := "run_analysis.bash"
	if app.config.DevelopmentMode {
		scriptName = "run_analysis_dev.bash"
	}
	args := fmt.Sprintf("bash %s %s %s", scriptName, eventLogPath, jobDir)

	cmd := exec.CommandContext(ctx, "sh", "-c", args)

	if runtime.GOOS != "windows" {
		// sets process group ID to kill all processes in the group later on cancel if needed
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	// capture stdout and stderr
	cmd.Stdout = app.logger.Writer()
	var buf bytes.Buffer
	errWriter := io.MultiWriter(app.logger.Writer(), &buf)
	cmd.Stderr = errWriter

	// interrupt the command if the context is cancelled
	go func() {
		select {
		case <-ctx.Done():
			if err = syscall.Kill(-1*cmd.Process.Pid, syscall.SIGKILL); err != nil {
				app.logger.Printf("Error cancelling job: %s", err.Error())
			}
		}
	}()

	if err = cmd.Start(); err != nil {
		return errors.New(fmt.Sprintf("error starting analysis: %s", err.Error()))
	}

	app.logger.Printf("Job %s executing", job.ID)

	if err = cmd.Wait(); err != nil {
		err = fmt.Errorf("error executing analysis: %s; stderr: %s", err.Error(), buf.String())
	}
	return err
}

func (app *Application) newJobFromRequestBody(body io.ReadCloser) (*model.Job, error) {
	defer func() {
		if err := body.Close(); err != nil {
			app.logger.Printf("error closing request body: %s", err.Error())
		}
	}()

	jobID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	jobDir := strings.Join([]string{app.config.ResultsDir, jobID.String()}, "/")

	logName := "event_log.csv"

	logPath := path.Join(jobDir, logName)

	if err := mkdir(jobDir); err != nil {
		return nil, err
	}

	f, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			app.logger.Printf("error closing file: %s", err.Error())
		}
	}()

	if _, err := io.Copy(f, body); err != nil {
		return nil, err
	}

	host := os.Getenv("WEBAPP_HOST")
	if len(host) == 0 {
		host = app.config.Host
	}

	eventLog := fmt.Sprintf("http://%s/assets/results/%s/%s", host, jobID, logName)

	eventLogURL, err := url.Parse(eventLog)
	if err != nil {
		return nil, err
	}

	return &model.Job{
		ID:                      jobID.String(),
		Status:                  model.JobStatusPending,
		EventLog:                eventLog,
		EventLogURL:             &model.URL{URL: eventLogURL},
		EventLogFromRequestBody: true,
		CreatedAt:               time.Now(),
		Dir:                     jobDir,
	}, nil
}
