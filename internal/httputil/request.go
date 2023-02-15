package httputil

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type CallOpts struct {
	Request    *http.Request
	Handler    ResponseHandler
	Validators []ResponseHandler
}

// Call calls the underlying http.Client and validates and handles any resulting response.
// The response body is closed after all validators and the handler run.
func Call(cli *http.Client, opts CallOpts) (err error) {
	res, err := cli.Do(opts.Request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if len(opts.Validators) == 0 {
		opts.Validators = []ResponseHandler{DefaultValidator}
	}
	if err = ChainHandlers(opts.Validators...)(res); err != nil {
		return err
	}

	handle := opts.Handler
	if handle == nil {
		handle = consumeBody
	}

	return handle(res)
}

// BasicAuth creates the Authorization header for basic auth credential.
func BasicAuth(username, password string) Multimap {
	auth := []byte(fmt.Sprintf("%s:%s", username, password))
	val := base64.StdEncoding.EncodeToString(auth)
	return Multimap{
		Key: "Authorization",
		Values: []string{
			fmt.Sprintf("Basic %s", val),
		},
	}
}

// Bearer creates the Authorization header for a bearer token.
func Bearer(token string) Multimap {
	return Multimap{
		Key: "Authorization",
		Values: []string{
			fmt.Sprintf("Bearer %s", token),
		},
	}
}

// Accept sets the Accept header for a request.
func Accept(contentTypes string) Multimap {
	return Multimap{
		Key: "Accept",
		Values: []string{
			contentTypes,
		},
	}
}

// ContentType sets the Content-Type header on a request.
func ContentType(ct string) Multimap {
	return Multimap{
		Key:    "Content-Type",
		Values: []string{ct},
	}
}

// UserAgent sets the User-Agent header.
func UserAgent(ua string) Multimap {
	return Multimap{
		Key:    "User-Agent",
		Values: []string{ua},
	}
}
