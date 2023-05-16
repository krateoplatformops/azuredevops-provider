package azuredevops

import (
	"net/http"

	"github.com/lucasepe/httplib"
)

type Resource struct {
	// Id of the resource.
	Id *string `json:"id,omitempty"`
	// Name of the resource.
	Name *string `json:"name,omitempty"`
	// Type of the resource.
	Type *string `json:"type,omitempty"`
}

func IsAlreadyExists(err error) bool {
	return httplib.HasStatusErr(err, http.StatusConflict)
}

func IsNotFound(err error) bool {
	return httplib.HasStatusErr(err, http.StatusNotFound)
}
