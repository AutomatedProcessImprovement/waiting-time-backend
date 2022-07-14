package app

import (
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"sort"
	"sync"
)

type Queue struct {
	Jobs []*model.Job

	lock sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{}
}

// Add adds a job to the queue if it's not yet present there.
func (q *Queue) Add(job *model.Job) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	for _, j := range q.Jobs {
		if j == job {
			return fmt.Errorf("job already present in queue")
		}
	}
	q.Jobs = append(q.Jobs, job)

	return nil
}

// Remove removes a job from the queue if it finds one.
func (q *Queue) Remove(job *model.Job) {
	if job == nil {
		return
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	for i, j := range q.Jobs {
		if j == job {
			q.Jobs = append(q.Jobs[:i], q.Jobs[i+1:]...)
			return
		}
	}
}

// FindByID finds a job by its ID.
func (q *Queue) FindByID(id string) *model.Job {
	for _, j := range q.Jobs {
		if j == nil {
			continue
		}

		if j.ID == id {
			return j
		}
	}
	return nil
}

// Next finds the first pending job in the queue.
func (q *Queue) Next() *model.Job {
	q.sort()

	for _, j := range q.Jobs {
		if j == nil {
			continue
		}

		if j.Status == model.JobStatusPending {
			return j
		}
	}
	return nil
}

func (q *Queue) sort() {
	q.lock.Lock()
	defer q.lock.Unlock()

	sort.Slice(q.Jobs, func(i, j int) bool {
		return q.Jobs[i].CreatedAt.Before(q.Jobs[j].CreatedAt)
	})
}
