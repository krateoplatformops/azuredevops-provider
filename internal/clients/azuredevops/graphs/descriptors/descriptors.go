package descriptors

import (
	"context"
	"net/http"
	"path"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
	"github.com/pkg/errors"
)

type DescriptorResponse struct {
	Links interface{} `json:"_links"`
	Value *string     `json:"value"`
}

// Options for the Get Pipeline Permissions ForResource function
type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	ResourceID   string
}

// Resolve a storage key to a descriptor
// GET https://vssps.dev.azure.com/{organization}/_apis/graph/descriptors/{storageKey}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*DescriptorResponse, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/descriptors", opts.ResourceID),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}

	res := &DescriptorResponse{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(res),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusOK),
		},
	})

	return res, err
}

func GetDescriptor(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*string, error) {
	res, err := Get(ctx, cli, opts)
	if res == nil {
		return nil, errors.Errorf("No descriptor for %s/%s", opts.Organization, opts.ResourceID)
	}
	return res.Value, err
}
