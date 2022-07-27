package model

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
