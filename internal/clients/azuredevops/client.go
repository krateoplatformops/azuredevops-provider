package azuredevops

import (
	"net/http"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httplib"
)

const (
	apiVersionKey = "api-version"
	apiVersionVal = "7.0"
	userAgent     = "krateo/azuredevops-provider"
)

type Options struct {
	BaseURL string
	Token   string
	Verbose bool
}

type Client struct {
	httpClient *http.Client
	options    Options
}

func NewClient(httpClient *http.Client, opts Options) *Client {
	return &Client{
		httpClient: httpClient,
		options:    opts,
	}
}

func (c *Client) newGetRequest(path string, queryParams map[string]string) (*http.Request, error) {
	if len(queryParams) == 0 {
		queryParams = map[string]string{}
	}
	queryParams[apiVersionKey] = apiVersionVal

	req, err := httplib.NewRequest(httplib.CreateRequestOpts{
		Method:      http.MethodGet,
		BaseURL:     c.options.BaseURL,
		Path:        path,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, err
	}
	//req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(userAgent, c.options.Token)
	return req, nil
}

func (c *Client) newPostRequest(path string, queryParams map[string]string, val any) (*http.Request, error) {
	if len(queryParams) == 0 {
		queryParams = map[string]string{}
	}
	queryParams[apiVersionKey] = apiVersionVal

	req, err := httplib.NewRequest(httplib.CreateRequestOpts{
		Method:      http.MethodPost,
		BaseURL:     c.options.BaseURL,
		Path:        path,
		QueryParams: queryParams,
		GetBody:     httplib.BodyJSON(val),
	})
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(userAgent, c.options.Token)
	return req, nil
}
