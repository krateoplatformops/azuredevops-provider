package projects

import (
	"errors"
	"fmt"
	"net/http"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httputil"
)

type APIError struct {
	Message   string `json:"message"`
	TypeKey   string `json:"typeKey"`
	ErrorCode int    `json:"errorCode"`
	EventID   int    `json:"eventId"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("azuredevops: %s (%s, %d)", e.Message, e.TypeKey, e.EventID)
}

// IsNotFound checks if the error has a not found status.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	se := &httputil.StatusError{}
	if errors.As(err, se) {
		return se.StatusCode == http.StatusNotFound
	}

	return false
}
