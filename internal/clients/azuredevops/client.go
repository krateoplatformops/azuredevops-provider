package azuredevops

import (
	"fmt"
	"net/http"

	"github.com/lucasepe/httplib"
)

const (
	ApiVersionKey  = "api-version"
	ApiVersionVal  = "7.0"
	ApiPreviewFlag = "-preview"
	UserAgent      = "krateo/azuredevops-provider"
)

type URIKey string

const (
	Default URIKey = "default"
	Feeds   URIKey = "feeds"
	Vssps   URIKey = "vssps"
)

type ClientOptions struct {
	Token   string
	Verbose bool
	UriMap  *map[URIKey]string
}

type Client struct {
	httpClient *http.Client
	uriMap     map[URIKey]string
	verbose    bool
	authMethod httplib.AuthMethod
}

func NewClient(opts ClientOptions) *Client {
	if opts.UriMap == nil {
		opts.UriMap = &map[URIKey]string{
			Default: "https://dev.azure.com",
			Feeds:   "https://feeds.dev.azure.com",
			Vssps:   "https://vssps.dev.azure.com",
		}
	}
	return &Client{
		httpClient: httplib.NewClient(),
		uriMap:     *opts.UriMap,
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

func (c *Client) BaseURL(loc URIKey) string {
	val, ok := c.uriMap[loc]
	if !ok {
		return c.uriMap[Default]
	}
	return val
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
