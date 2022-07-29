package model

// ApiCallbackRequest is a body for POST request to the callback endpoint that was specified during job submission.
//
// swagger:model
type ApiCallbackRequest struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
