package model

// ApiResponseError represents an error response from the API.
//
// swagger:model
type ApiResponseError struct {
	Message string `json:"message,omitempty"`
}
