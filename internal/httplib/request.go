package httplib

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type CreateRequestOpts struct {
	Method      string
	BaseURL     string
	Path        string
	QueryParams map[string]string

	// Headers for the http request.
	Headers map[string]string

	// BodyGetter provides a source for a request body.
	GetBody func() (io.ReadCloser, error)
}

func NewRequest(opts CreateRequestOpts) (*http.Request, error) {
	base, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, err
	}
	uri := base.JoinPath(opts.Path)
	if err != nil {
		return nil, err
	}
	if len(opts.QueryParams) > 0 {
		q := uri.Query()
		for k, v := range opts.QueryParams {
			q.Add(k, v)
		}
		uri.RawQuery = q.Encode()
	}

	var body io.Reader
	if opts.GetBody != nil {
		if body, err = opts.GetBody(); err != nil {
			return nil, err
		}
		if nopper, ok := body.(nopCloser); ok {
			body = nopper.Reader
		}
	}

	req, err := http.NewRequest(opts.Method, uri.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.GetBody = opts.GetBody

	for k, v := range opts.Headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

type CallOpts struct {
	Verbose         bool
	ResponseHandler ResponseHandler
	Validators      []ResponseHandler
}

// Call calls the underlying http.Client and validates and handles any resulting response.
// The response body is closed after all validators and the handler run.
func Call(httpClient *http.Client, req *http.Request, opts CallOpts) (err error) {
	if opts.Verbose {
		// Dump the request to os.Stderr.
		buf, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			os.Stderr.Write(buf)
			os.Stderr.Write([]byte{'\n'})
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if opts.Verbose {
		// Dump the response to os.Stderr.
		buf, err := httputil.DumpResponse(res, req.URL.Query().Get("watch") != "true")
		if err == nil {
			os.Stderr.Write(buf)
			os.Stderr.Write([]byte{'\n'})
		}
	}

	if len(opts.Validators) == 0 {
		opts.Validators = []ResponseHandler{
			CheckStatus(
				http.StatusOK,
				http.StatusCreated,
				http.StatusAccepted,
				http.StatusNonAuthoritativeInfo,
				http.StatusNoContent,
			),
		}
	}
	err = ChainHandlers(opts.Validators...)(res)
	if err != nil {
		return err
	}

	handle := opts.ResponseHandler
	if handle == nil {
		handle = consumeBody
	}

	return handle(res)
}
