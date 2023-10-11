package app

import (
	"os"
	"testing"
	"time"

	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
)

func TestNewApplication(t *testing.T) {
	config := &Configuration{
		Port:           8080,
		QueueSleepTime: time.Second * 10,
		JobTimeout:     time.Minute * 5,
		ResultsDir:     "./results/",
		QueuePath:      "./queue.gob",
		Host:           "localhost",
	}

	app, err := NewApplication(config)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if app == nil {
		t.Fatal("Expected application to be instantiated, but got nil")
	}
}

func TestAddJob(t *testing.T) {
	config := &Configuration{
		Port:           8080,
		QueueSleepTime: time.Second * 10,
		JobTimeout:     time.Minute * 5,
		ResultsDir:     "./results/",
		QueuePath:      "./queue.gob",
		Host:           "localhost",
	}

	app, _ := NewApplication(config)

	job := &model.Job{
		ID: "test-job-id",
	}

	err := app.AddJob(job)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
