package app

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type JobStatus string

var (
	JobStatusPending   = JobStatus("pending")
	JobStatusRunning   = JobStatus("running")
	JobStatusCompleted = JobStatus("completed")
	JobStatusFailed    = JobStatus("failed")
)

type Job struct {
	ID               string     `json:"id,omitempty"`
	Status           JobStatus  `json:"status,omitempty"`
	Result           *JobResult `json:"result,omitempty"`
	ReportCSV        *URL       `json:"report_csv,string,omitempty"`
	CallbackEndpoint *URL       `json:"callback_endpoint,string,omitempty"`
	EventLog         *URL       `json:"event_log,string,omitempty"`
	CreatedAt        time.Time  `json:"created_at,omitempty"`
	CompletedAt      *time.Time `json:"finished_at,omitempty"`
}

func NewJob(eventLog *URL, callback *URL) (*Job, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return &Job{
		ID:               id.String(),
		Status:           JobStatusPending,
		CallbackEndpoint: callback,
		EventLog:         eventLog,
		CreatedAt:        time.Now(),
	}, nil
}

func (j *Job) Validate() error {
	if j.ID == "" {
		return fmt.Errorf("job ID is required")
	}

	if j.Status == "" {
		return fmt.Errorf("job status is required")
	}

	if j.CallbackEndpoint.String() == "" {
		return fmt.Errorf("job callback endpoint is required")
	}

	if j.EventLog.String() == "" {
		return fmt.Errorf("job event log is required")
	}

	return nil
}
