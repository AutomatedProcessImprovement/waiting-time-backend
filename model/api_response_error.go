package model

// ApiResponseError represents an error response from the API.
//
// swagger:model
type ApiResponseError struct {
	Error string `json:"error,omitempty"`
}
