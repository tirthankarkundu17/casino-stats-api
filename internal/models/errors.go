package models

import "fmt"

// ErrNotFound is returned when a requested resource does not exist.
type ErrNotFound struct {
	Resource string
	Message  string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s: %s", e.Resource, e.Message)
}

// ErrNoData is returned when a query returns no results for the given filters.
type ErrNoData struct {
	Message string
}

func (e *ErrNoData) Error() string {
	return e.Message
}

// APIError represents a structured error response
type APIError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail provides context for specific field failures
type ErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}
