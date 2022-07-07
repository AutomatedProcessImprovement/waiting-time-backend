package app

import "fmt"

type Queue struct {
	Jobs []*Job `json:"jobs,omitempty"`
}

func NewQueue() *Queue {
	return &Queue{}
}

// Add adds a job to the queue if it's not yet present there.
func (q *Queue) Add(job *Job) error {
	for _, j := range q.Jobs {
		if j == job {
			return fmt.Errorf("job already present in queue")
		}
	}
	q.Jobs = append(q.Jobs, job)

	return nil
}

// Remove removes a job from the queue if it finds one.
func (q *Queue) Remove(job *Job) {
	for i, j := range q.Jobs {
		if j == job {
			q.Jobs = append(q.Jobs[:i], q.Jobs[i+1:]...)
			return
		}
	}
}

// FindByID finds a job by its ID.
func (q *Queue) FindByID(id string) *Job {
	for _, j := range q.Jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}
