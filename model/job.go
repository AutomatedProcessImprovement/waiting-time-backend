package model

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
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

type Job struct {
	ID               string     `json:"id,omitempty"`
	Status           JobStatus  `json:"status,omitempty"`
	Error            string     `json:"error,omitempty"`
	Result           *JobResult `json:"result,omitempty"`
	ReportCSV        *URL       `json:"report_csv,string,omitempty"`
	CallbackEndpoint *URL       `json:"callback_endpoint,string,omitempty"`
	EventLog         *URL       `json:"event_log,string,omitempty"`
	CreatedAt        time.Time  `json:"created_at,omitempty"`
	CompletedAt      *time.Time `json:"finished_at,omitempty"`

	lock sync.Mutex
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

func (j *Job) Run(logger *log.Logger) {
	var err error

	if j.Status != JobStatusPending {
		err = fmt.Errorf("job is not pending")
		logger.Printf("Job %s failed; %s", j.ID, err.Error())
		j.Error = err.Error()
		return
	}

	// pre-work
	j.lock.Lock()
	j.Status = JobStatusRunning
	j.lock.Unlock()

	logger.Printf("Job %s started", j.ID)

	// useful work
	jobDuration := time.Duration(rand.Intn(60)+60) * time.Second
	time.Sleep(jobDuration)

	// post-work
	j.lock.Lock()
	defer j.lock.Unlock()

	now := time.Now()
	j.CompletedAt = &now

	if err == nil {
		logger.Printf("Job %s completed", j.ID)
		j.Status = JobStatusCompleted
	} else {
		logger.Printf("Job %s failed; %s", j.ID, err.Error())
		j.Status = JobStatusFailed
		j.Error = err.Error()
	}
}
