package azuredevops

import (
	"net/http"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httputil"
)

const (
	UserAgent = "krateo"
)

var (
	APIVersion = httputil.NewMultimap("api-version", "7.0")
)

type ClientOpts struct {
	Verbose  bool
	Insecure bool
	BaseURL  string
	Token    string
}

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(opts ClientOpts) *Client {
	return &Client{
		BaseURL: opts.BaseURL,
		Token:   opts.Token,
		HTTPClient: httputil.ClientFromOpts(httputil.ClientOpts{
			Verbose:  opts.Verbose,
			Insecure: opts.Insecure,
		}),
	}
}
