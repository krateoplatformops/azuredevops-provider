package httplib

import (
	"bytes"
	"encoding/json"
	"io"
)

// nopCloser is like io.NopCloser(),
// but it is a concrete type so we can strip it out
// before setting a body on a request.
// See https://github.com/carlmjohnson/requests/discussions/49
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

var _ io.ReadCloser = nopCloser{}

// BodyGetter provides a Builder with a source for a request body.
type BodyGetter = func() (io.ReadCloser, error)

// BodyJSON is a BodyGetter that marshals a JSON object.
func BodyJSON(v any) BodyGetter {
	return func() (io.ReadCloser, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return nopCloser{bytes.NewReader(b)}, nil
	}
}
