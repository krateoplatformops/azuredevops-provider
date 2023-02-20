package azuredevops

import (
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

/*
type urlBuilder struct {
	baseURL string
	path string
	params []string
}

var _ httplib.URLBuilder = (*urlBuilder)(nil)

func (ub *urlBuilder) Build() (*url.URL, error) {

}
*/
/*
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

func (c *Client) newDeleteRequest(path string, queryParams map[string]string) (*http.Request, error) {
	if len(queryParams) == 0 {
		queryParams = map[string]string{}
	}
	queryParams[apiVersionKey] = apiVersionVal

	req, err := httplib.NewRequest(httplib.CreateRequestOpts{
		Method:      http.MethodDelete,
		BaseURL:     c.options.BaseURL,
		Path:        path,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(userAgent, c.options.Token)
	return req, nil
}
*/
