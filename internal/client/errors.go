package client

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned for HTTP 404 responses. Use errors.Is(err, ErrNotFound)
// to detect a missing resource.
var ErrNotFound = errors.New("hudu: resource not found")

// APIError carries an error response from the Hudu API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("hudu: %s %s returned HTTP %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

func (e *APIError) Is(target error) bool {
	if target == ErrNotFound {
		return e.StatusCode == 404
	}
	return false
}
