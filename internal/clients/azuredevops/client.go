package azuredevops

import (
	"fmt"
	"net/http"

	"github.com/lucasepe/httplib"
)

const (
	apiVersionKey = "api-version"
	apiVersionVal = "7.0"
	userAgent     = "krateo/azuredevops-provider"
)

type ClientOptions struct {
	BaseURL string
	Token   string
	Verbose bool
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	verbose    bool
	authMethod httplib.AuthMethod
}

func NewClient(opts ClientOptions) *Client {
	return &Client{
		httpClient: httplib.NewClient(),
		baseURL:    opts.BaseURL,
		verbose:    opts.Verbose,
		authMethod: &httplib.BasicAuth{
			Username: userAgent,
			Password: opts.Token,
		},
	}
}

type APIError struct {
	Message   string `json:"message"`
	TypeKey   string `json:"typeKey"`
	ErrorCode int    `json:"errorCode"`
	EventID   int    `json:"eventId"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("azuredevops: %s (%s, %d)", e.Message, e.TypeKey, e.EventID)
}
