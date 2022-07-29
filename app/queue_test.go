package app

import (
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestQueue_Clear(t *testing.T) {
	const resultsDir = "../assets/results"
	var rootFS = os.DirFS(resultsDir)

	j, err := model.NewJob(nil, nil, resultsDir)
	if err != nil {
		t.Fatal(err)
	}

	q := NewQueue()
	if err = q.Add(j); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		queue   *Queue
		wantErr bool
	}{
		{
			name:    "valid queue",
			queue:   q,
			wantErr: false,
		},
		{
			name: "valid queue with invalid job",
			queue: &Queue{
				Jobs: []*model.Job{
					nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.queue.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("Clear() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.queue != nil && len(tt.queue.Jobs) != 0 {
				t.Errorf("Clear() jobs = %v, want %v", len(tt.queue.Jobs), 0)
			}

			matches, err := fs.Glob(rootFS, "*")
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != 0 {
				t.Errorf("Clear() files = %v, want %v", len(matches), 0)
			}
		})
	}
}

func TestQueue_ClearOld(t *testing.T) {
	const resultsDir = "../assets/results"
	var rootFS = os.DirFS(resultsDir)

	j := &model.Job{
		ID:               "1",
		Status:           model.JobStatusPending,
		CallbackEndpoint: "",
		EventLog:         "",
		CreatedAt:        time.Now().Add(-1 * 24 * 2 * time.Hour),
		CompletedAt:      nil,
		Dir:              "",
	}

	j2 := &model.Job{
		ID:               "2",
		Status:           model.JobStatusPending,
		CallbackEndpoint: "",
		EventLog:         "",
		CreatedAt:        time.Now(),
		CompletedAt:      nil,
		Dir:              "",
	}

	q := NewQueue()
	if err := q.Add(j); err != nil {
		t.Fatal(err)
	}

	q2 := NewQueue()
	if err := q2.Add(j); err != nil {
		t.Fatal(err)
	}
	if err := q2.Add(j2); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name              string
		queue             *Queue
		expectedJobsCount int
		wantErr           bool
	}{
		{
			name:    "valid queue",
			queue:   q,
			wantErr: false,
		},
		{
			name:    "nil queue",
			queue:   nil,
			wantErr: true,
		},
		{
			name: "valid queue with invalid job",
			queue: &Queue{
				Jobs: []*model.Job{
					nil, // invalid job with no CreatedAt date
				},
			},
			expectedJobsCount: 1,
			wantErr:           false,
		},
		{
			name:              "queue with old and new jobs",
			queue:             q2,
			expectedJobsCount: 1,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.queue.ClearOld(-1 * 6 * time.Hour); (err != nil) != tt.wantErr {
				t.Fatalf("ClearOld() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.queue != nil && len(tt.queue.Jobs) != tt.expectedJobsCount && !tt.wantErr {
				t.Fatalf("ClearOld() jobs = %v, want %v", len(tt.queue.Jobs), tt.expectedJobsCount)
			}

			matches, err := fs.Glob(rootFS, "*")
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != 0 {
				t.Errorf("ClearOld() files = %v, want %v", len(matches), 0)
			}
		})
	}
}
