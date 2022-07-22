package model

// swagger:model
type ApiResponse struct {
	Job    *Job              `json:"job,omitempty"`
	Jobs   []*Job            `json:"jobs,omitempty"`
	Error_ *ApiResponseError `json:"error,omitempty"`
}

// ApiSingleJobResponse is a response for a single job operation.
//
// swagger:model
type ApiSingleJobResponse struct {
	*Job
}

// ApiJobsResponse is a response for multiple jobs operation.
//
// swagger:model
type ApiJobsResponse struct {
	Jobs []*Job `json:"jobs"`
}
