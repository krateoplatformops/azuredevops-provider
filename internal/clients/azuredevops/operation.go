package azuredevops

import (
	"context"
	"net/http"
	"path"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httplib"
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

// Gets an operation from the operationId using the given pluginId.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/operations/operations/get?view=azure-devops-rest-7.0#operation
func GetOperation(ctx context.Context, cli *Client, opts GetOperationOpts) (*Operation, error) {
	apiPath := path.Join(opts.Organization, "_apis/operations", opts.OperationId)
	req, err := cli.newGetRequest(apiPath, nil)
	if err != nil {
		return nil, err
	}

	apiErr := &APIError{}
	val := &Operation{}

	err = httplib.Call(cli.httpClient, req, httplib.CallOpts{
		Verbose:         cli.options.Verbose,
		ResponseHandler: httplib.ToJSON(val),
		Validators: []httplib.ResponseHandler{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}
