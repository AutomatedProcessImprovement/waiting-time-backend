package app

type ApiResponse struct {
	Job    *Job              `json:"job,omitempty"`
	Jobs   []*Job            `json:"jobs,omitempty"`
	Error_ *ApiResponseError `json:"error,omitempty"`
}
