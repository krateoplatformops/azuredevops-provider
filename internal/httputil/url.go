package httputil

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
)

// BuildURL builds a *url.URL from the base URL and options set.
// If a valid url.URL cannot be built,
// URL() nevertheless returns a new url.URL,
// so it is always safe to call u.String().
func BuildURL(baseurl string, paths []string, params ...Multimap) (u *url.URL, err error) {
	u, err = url.Parse(baseurl)
	if err != nil {
		return new(url.URL), err
	}

	u.Path = u.ResolveReference(&url.URL{Path: path.Join(paths...)}).Path

	if len(params) > 0 {
		q := u.Query()
		for _, kv := range params {
			q[kv.Key] = kv.Values
		}
		u.RawQuery = q.Encode()
	}

	// Reparsing, in case the path rewriting broke the URL
	u, err = url.Parse(u.String())
	if err != nil {
		return new(url.URL), err
	}
	return u, nil
}

type RequestOpts struct {
	// URL is the the base url.
	URL string

	// Method is the request method.
	Method string

	// Cookies for request.
	Cookies []KeyValPair

	// Headers for the http request.
	Headers []Multimap

	// BodyGetter provides a source for a request body.
	GetBody func() (io.ReadCloser, error)
}

// Request builds a new http.Request with its context set.
func Request(ctx context.Context, opts RequestOpts) (req *http.Request, err error) {
	var body io.Reader
	if opts.GetBody != nil {
		if body, err = opts.GetBody(); err != nil {
			return nil, err
		}
		if nopper, ok := body.(nopCloser); ok {
			body = nopper.Reader
		}
	}

	method := http.MethodGet
	if len(opts.Method) > 0 {
		method = opts.Method
	}

	req, err = http.NewRequestWithContext(ctx, method, opts.URL, body)
	if err != nil {
		return nil, err
	}
	req.GetBody = opts.GetBody

	for _, kv := range opts.Headers {
		req.Header[http.CanonicalHeaderKey(kv.Key)] = kv.Values
	}

	for _, kv := range opts.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  kv.Key,
			Value: kv.Value,
		})
	}

	return req, nil
}

func NewMultimap(key string, values ...string) Multimap {
	return Multimap{key, values}
}
