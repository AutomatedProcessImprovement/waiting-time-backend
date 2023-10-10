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
//   - application/json
//
// Produces:
//   - application/json
//
// swagger:meta
package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
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

// ProcessQueue should be started in a separate goroutine to run the queue processing alongside the web server.
// It checks for any pending job and process one if available. It also saves the queue to disk when processing is done,
// and clears old records periodically.
func (app *Application) ProcessQueue() {
	app.logger.Printf("Queue processing started")

	for {
		// empties queue and disk monthly
		if err := app.queue.ClearOld(-24 * 31 * time.Hour); err != nil {
			app.logger.Printf("Error clearing old jobs: %s", err.Error())
		}

		// checks if there are jobs in the queue
		job := app.queue.Next()
		if job == nil {
			time.Sleep(app.config.QueueSleepTime)
			continue
		}

		// executes the job and saves the result on disk
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
	// post-work
	defer func() {
		job.SetCompletedAt(time.Now())

		if err := app.callback(job); err != nil {
			app.logger.Printf("Error calling callback endpoint for job %s: %s", job.ID, err.Error())
			job.SetError(err)
		}
	}()

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
		if foundJob != nil && foundJob.ID != job.ID &&
			foundJob.Status != model.JobStatusPending &&
			foundJob.Status != model.JobStatusFailed { // allow re-processing of failed jobs
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
				_, err = app.prepareJobResult(job)
				if err != nil {
					app.logger.Printf("error preparing result: %s", err.Error())
					job.SetError(err)
				} else {
					//job.SetResult(result)
				}
			}
		}
	}
}

func (app *Application) callback(job *model.Job) error {
	if job.CallbackEndpointURL == nil {
		return nil
	}

	payload := model.ApiCallbackRequest{
		JobID:  job.ID,
		Status: fmt.Sprintf("%s", job.Status),
		Error:  job.Error,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	r := bytes.NewReader(buf.Bytes())

	req, err := http.NewRequest("POST", job.CallbackEndpointURL.String(), r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	_, err = http.DefaultClient.Do(req)

	return err
}

func (app *Application) prepareJobResult(job *model.Job) ([]model.JobResultItem, error) {
	if job.EventLogURL == nil {
		return nil, fmt.Errorf("job has no event log")
	}

	const (
		reportSuffixCSV = "_transitions_report.csv"
	)

	eventLogName := path.Base(job.EventLogURL.String())
	eventLogExt := path.Ext(eventLogName)
	resultName := strings.TrimSuffix(eventLogName, eventLogExt) + reportSuffixCSV
	resultPath := path.Join(job.Dir, resultName)

	results, err := app.jobResultsFromPath(resultPath)
	if err != nil {
		return nil, fmt.Errorf("error reading result: %s", err.Error())
	}

	// Save results to database
	err = app.storeJobResultsInDatabase(job.ID, results)
	if err != nil {
		return nil, fmt.Errorf("error storing results in database: %s", err.Error())
	}

	return results, nil
}

func (app *Application) jobResultsFromPath(filePath string) ([]model.JobResultItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 1 {
		return nil, errors.New("empty CSV")
	}

	var results []model.JobResultItem
	timeLayout := "2006-01-02 15:04:05-07:00"

	for _, record := range records[1:] {
		startTime, _ := time.Parse(timeLayout, record[0])
		endTime, _ := time.Parse(timeLayout, record[1])
		wtTotal, _ := strconv.ParseFloat(record[7], 64)
		wtContention, _ := strconv.ParseFloat(record[8], 64)
		wtBatching, _ := strconv.ParseFloat(record[9], 64)
		wtPrioritization, _ := strconv.ParseFloat(record[10], 64)
		wtUnavailability, _ := strconv.ParseFloat(record[11], 64)
		wtExtraneous, _ := strconv.ParseFloat(record[12], 64)

		result := model.JobResultItem{
			StartTime:           startTime,
			EndTime:             endTime,
			SourceActivity:      record[2],
			SourceResource:      record[3],
			DestinationActivity: record[4],
			DestinationResource: record[5],
			CaseID:              record[6],
			WtTotal:             wtTotal,
			WtContention:        wtContention,
			WtBatching:          wtBatching,
			WtPrioritization:    wtPrioritization,
			WtUnavailability:    wtUnavailability,
			WtExtraneous:        wtExtraneous,
		}
		results = append(results, result)
	}

	return results, nil
}

func sanitizeTableName(input string) string {
	// This regex matches characters that are not alphanumeric
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(input, "_")
}

func (app *Application) storeJobResultsInDatabase(jobID string, results []model.JobResultItem) error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return errors.New("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create table
	tableName := "result_" + sanitizeTableName(jobID)
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE %s (
            StartTime TIMESTAMP,
            EndTime TIMESTAMP,
            SourceActivity TEXT,
            SourceResource TEXT,
            DestinationActivity TEXT,
            DestinationResource TEXT,
            CaseID TEXT,
            WtTotal FLOAT,
            WtContention FLOAT,
            WtBatching FLOAT,
            WtPrioritization FLOAT,
            WtUnavailability FLOAT,
            WtExtraneous FLOAT
        )
    `, tableName)

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	for _, result := range results {
		insertSQL := fmt.Sprintf(`
            INSERT INTO %s (
                StartTime, EndTime, SourceActivity, SourceResource,
                DestinationActivity, DestinationResource, CaseID,
                WtTotal, WtContention, WtBatching, WtPrioritization,
                WtUnavailability, WtExtraneous
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        `, tableName)

		_, err = db.Exec(insertSQL,
			result.StartTime, result.EndTime, result.SourceActivity, result.SourceResource,
			result.DestinationActivity, result.DestinationResource, result.CaseID,
			result.WtTotal, result.WtContention, result.WtBatching, result.WtPrioritization,
			result.WtUnavailability, result.WtExtraneous,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) newJobFromRequestBody(body io.ReadCloser, columnMapping map[string]string) (*model.Job, error) {
	defer func() {
		if err := body.Close(); err != nil {
			app.logger.Printf("error closing request body: %s", err.Error())
		}
	}()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %s", err.Error())
	}
	if len(bodyBytes) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}

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

	if _, err := io.Copy(f, bytes.NewReader(bodyBytes)); err != nil {
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
		ColumnMapping:           columnMapping,
	}, nil
}
