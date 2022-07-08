package app

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type Application struct {
	router        *mux.Router
	queue         *Queue
	config        *Configuration
	docker        *DockerWorker
	logger        *log.Logger
	loggerFile    *os.File
	webLogger     *log.Logger
	webLoggerFile *os.File
}

func NewApplication(config *Configuration) (*Application, error) {
	app := &Application{
		config: config,
		queue:  NewQueue(),
		docker: DefaultDockerWorker(),
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
	app.loggerFile = appFile
	app.logger = log.New(appFile, "", log.Ldate|log.Ltime)

	requestsFile, err := os.OpenFile(config.WebLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	app.webLoggerFile = requestsFile
	app.webLogger = log.New(requestsFile, "", log.Ldate|log.Ltime)

	if err := mkdir(config.ResultsDir); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *Application) Close() {
	if err := app.loggerFile.Close(); err != nil {
		app.logger.Printf("error closing file: %s", err.Error())
	}
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
			jobErrorChan <- app.runAnalysis(ctx, eventLogName, job.Dir)
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
				reportURL, err := url.Parse(fmt.Sprintf("http://localhost:8080/assets/results/%s/%s", job.ID, reportName))
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

func (app *Application) runAnalysis(ctx context.Context, eventLogPath, hostOutDir string) error {
	const containerDir = "/usr/src/app/result"
	logPathInContainer := fmt.Sprintf("%s/%s", containerDir, eventLogPath)

	hostOutDir, err := abspath(hostOutDir)
	if err != nil {
		return err
	}

	cmd := app.composeDockerCmd(hostOutDir, containerDir, logPathInContainer)

	out, err := app.docker.Run(ctx, app.logger, cmd)

	if err != nil {
		return err
	}

	if len(out) > 0 {
		app.logger.Printf("Docker output: %s", out)
	}

	return nil
}

func (app *Application) composeDockerCmd(hostOutDir string, containerDir string, logPathInContainer string) string {
	cmd := fmt.Sprintf("docker run --rm --name %s", app.docker.Container)

	if app.docker.WorkDir != "" {
		cmd += fmt.Sprintf(" -w %s", app.docker.WorkDir)
	}

	if len(app.docker.Env) > 0 {
		cmd += " -e " + strings.Join(app.docker.Env, " -e ")
	}

	cmd += fmt.Sprintf(" -v %s:%s", hostOutDir, containerDir)

	cmd += fmt.Sprintf(" %s", app.docker.Image)

	cmd += fmt.Sprintf(" bash run_analysis.sh %s %s", logPathInContainer, containerDir)
	return cmd
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

func mkdir(path string) error {
	if _, err := os.Open(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0777)
	}
	return nil
}

func abspath(p string) (string, error) {
	if !path.IsAbs(p) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		p = path.Join(cwd, p)
	}
	return p, nil
}

func download(url string, path string, logger *log.Logger) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Printf("error closing response body: %s", err.Error())
		}
	}()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			logger.Printf("error closing file: %s", err.Error())
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return err
}
