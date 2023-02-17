package azuredevops

import (
	"context"
	"net/http"
	"path"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httplib"
)

type OperationResultReference struct {
	// URL to the operation result.
	ResultUrl *string `json:"resultUrl,omitempty"`
}

// Contains information about the progress or result of an async operation.
type Operation struct {
	// Unique identifier for the operation.
	Id *string `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *string `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Links to other related objects.
	Links any `json:"_links,omitempty"`
	// Detailed messaged about the status of an operation.
	DetailedMessage *string `json:"detailedMessage,omitempty"`
	// Result message for an operation.
	ResultMessage *string `json:"resultMessage,omitempty"`
	// URL to the operation result.
	ResultUrl *OperationResultReference `json:"resultUrl,omitempty"`
}

// Reference for an async operation.
type OperationReference struct {
	// Unique identifier for the operation.
	Id *string `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *string `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
}

// The status of an operation.
type OperationStatus string

const (
	StatusNotSet     OperationStatus = "notSet"
	StatusQueued     OperationStatus = "queued"
	StatusInProgress OperationStatus = "inProgress"
	StatusCancelled  OperationStatus = "cancelled"
	StatusSucceeded  OperationStatus = "succeded"
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
