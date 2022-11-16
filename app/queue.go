package app

import (
	"fmt"
	"github.com/AutomatedProcessImprovement/waiting-time-backend/model"
	"os"
	"sort"
	"sync"
	"time"
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

// Remove removes a job from the queue if it finds one. It also can remove files related to the job.
func (q *Queue) Remove(job *model.Job, removeFiles bool) error {
	if job == nil {
		return nil
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	var newJobs []*model.Job
	for _, j := range q.Jobs {
		if j.ID == job.ID {
			continue
		}
		newJobs = append(newJobs, j)
	}
	q.Jobs = newJobs

	if removeFiles {
		if job.Dir == "" {
			return nil
		}

		return os.RemoveAll(job.Dir)
	}

	return nil
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

// FindByMD5 finds a job by the event log's MD5 hash. Returns nil if not found.
func (q *Queue) FindByMD5(md5 string) *model.Job {
	for _, j := range q.Jobs {
		if j == nil {
			continue
		}

		if j.EventLogMD5 == md5 {
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

// Clear empties the queue and removes related disk data.
func (q *Queue) Clear() error {
	q.lock.Lock()
	defer q.lock.Unlock()

	runningJobsCount := q.countRunningJobs()
	if q.countRunningJobs() > 0 {
		return fmt.Errorf("cannot clear queue while there are %d running jobs", runningJobsCount)
	}

	for _, j := range q.Jobs {
		if j == nil {
			continue
		}

		if j.Dir == "" {
			continue
		}

		if err := os.RemoveAll(j.Dir); err != nil {
			return fmt.Errorf("cannot remove job's [%s] folder: %s", j.ID, err)
		}
	}

	q.Jobs = []*model.Job{}

	return nil
}

// ClearOld removes jobs older than a given time from the queue and disk. Given duration should be negative to represent
// a time in the past.
func (q *Queue) ClearOld(d time.Duration) error {
	if q == nil {
		return fmt.Errorf("queue is nil")
	}

	for _, j := range q.Jobs {
		if j == nil {
			continue
		}

		if j.CreatedAt.Before(time.Now().Add(d)) {
			if err := q.Remove(j, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (q *Queue) countRunningJobs() int {
	runningJobsCount := 0

	for _, j := range q.Jobs {
		if j == nil {
			continue
		}

		if j.Status == model.JobStatusRunning {
			runningJobsCount++
		}
	}
	return runningJobsCount
}

func (q *Queue) sort() {
	q.lock.Lock()
	defer q.lock.Unlock()

	sort.Slice(q.Jobs, func(i, j int) bool {
		return q.Jobs[i].CreatedAt.Before(q.Jobs[j].CreatedAt)
	})
}
