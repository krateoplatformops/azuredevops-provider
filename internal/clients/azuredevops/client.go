package azuredevops

import (
	"fmt"
	"net/http"

	"github.com/lucasepe/httplib"
)

const (
	ApiVersionKey = "api-version"
	ApiVersionVal = "7.0"
	UserAgent     = "krateo/azuredevops-provider"
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
			Username: UserAgent,
			Password: opts.Token,
		},
	}
}

func (c *Client) SetVerbose(v bool) {
	c.verbose = v
}

func (c *Client) Verbose() bool {
	return c.verbose
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) AuthMethod() httplib.AuthMethod {
	return c.authMethod
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
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
