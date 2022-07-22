package model

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"sync"
	"time"
)

type JobStatus string

var (
	JobStatusPending   = JobStatus("pending")
	JobStatusRunning   = JobStatus("running")
	JobStatusCompleted = JobStatus("completed")
	JobStatusFailed    = JobStatus("failed")
)

// Job represents a job to be executed.
//
// swagger:model
type Job struct {
	ID                  string     `json:"id,omitempty"`
	Status              JobStatus  `json:"status,omitempty"`
	Error               string     `json:"error,omitempty"`
	Result              *JobResult `json:"result,omitempty"`
	ReportCSV           *URL       `json:"report_csv,omitempty"`
	CallbackEndpoint    string     `json:"callback_endpoint,omitempty"`
	CallbackEndpointURL *URL       `json:"-,omitempty"`
	EventLog            string     `json:"event_log,omitempty"`
	EventLogURL         *URL       `json:"-,omitempty"`
	CreatedAt           time.Time  `json:"created_at,omitempty"`
	CompletedAt         *time.Time `json:"finished_at,omitempty"`

	lock sync.Mutex
	Dir  string `json:"-"`
}

func NewJob(eventLog *URL, callback *URL, basedir string) (*Job, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:                  id.String(),
		Status:              JobStatusPending,
		CallbackEndpoint:    callback.String(),
		CallbackEndpointURL: callback,
		EventLog:            eventLog.String(),
		EventLogURL:         eventLog,
		CreatedAt:           time.Now(),
		Dir:                 strings.Join([]string{basedir, id.String()}, "/"),
	}, nil
}

func (j *Job) Validate() error {
	if j.ID == "" {
		return fmt.Errorf("job ID is required")
	}

	if j.Status == "" {
		return fmt.Errorf("job status is required")
	}

	if j.CallbackEndpointURL.String() == "" || j.CallbackEndpoint == "" {
		return fmt.Errorf("job callback endpoint is required")
	}

	if j.EventLogURL.String() == "" || j.EventLog == "" {
		return fmt.Errorf("job event log is required")
	}

	return nil
}

func (j *Job) SetStatus(status JobStatus) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.Status = status
}

func (j *Job) SetError(err error) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.Error = err.Error()
}

func (j *Job) SetResult(result *JobResult) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.Result = result
}

func (j *Job) SetReportCSV(url *URL) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.ReportCSV = url
}

func (j *Job) SetCompletedAt(t time.Time) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.CompletedAt = &t
}
