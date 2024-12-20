package azuredevops

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

// Contains information about the progress or result of an async operation.
type Operation struct {
	// Unique identifier for the operation.
	Id string `json:"id,omitempty"`
	// The current status of the operation.
	Status OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url string `json:"url,omitempty"`
	// Detailed messaged about the status of an operation.
	DetailedMessage string `json:"detailedMessage,omitempty"`
	// Result message for an operation.
	ResultMessage string `json:"resultMessage,omitempty"`
}

// Reference for an async operation.
type OperationReference struct {
	// Unique identifier for the operation.
	Id string `json:"id,omitempty"`
	// The current status of the operation.
	Status OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url string `json:"url,omitempty"`
}

// The status of an operation.
type OperationStatus string

const (
	StatusNotSet     OperationStatus = "notSet"
	StatusQueued     OperationStatus = "queued"
	StatusInProgress OperationStatus = "inProgress"
	StatusCancelled  OperationStatus = "cancelled"
	StatusSucceeded  OperationStatus = "succeeded"
	StatusFailed     OperationStatus = "failed"
)

type GetOperationOpts struct {
	Organization string
	OperationId  string
}

func getOperationsAPIVersion(cli *Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.Operations
		if apiVersion != nil {
			if strings.EqualFold(*apiVersion, "none") {
				apiVersionParams = nil
				isNone = true
			} else {
				apiVersionParams = []string{ApiVersionKey, helpers.String(apiVersion)}
			}
		}
	}
	return apiVersionParams, isNone
}

// Gets an operation from the operationId using the given pluginId.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/operations/operations/get?view=azure-devops-rest-7.0#operation
func (c *Client) GetOperation(ctx context.Context, opts GetOperationOpts) (*Operation, error) {
	apiVersionParams, isNone := getOperationsAPIVersion(c)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{ApiVersionKey, ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.BaseURL(Default),
		Path:    path.Join(opts.Organization, "_apis/operations", opts.OperationId),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &Operation{}

	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		Verbose:         c.verbose,
		AuthMethod:      c.authMethod,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}
