package projects

import (
	"context"
	"net/http"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httputil"
)

type GetProjectOpts struct {
	Organization string
	ProjectId    string
}

// Get project with the specified id or name, optionally including capabilities.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/core/projects/get?view=azure-devops-rest-7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetProjectOpts) (*TeamProject, error) {
	url, err := httputil.BuildURL(cli.BaseURL, []string{
		opts.Organization,
		"_apis/projects",
		opts.ProjectId,
	}, azuredevops.APIVersion)
	if err != nil {
		return nil, err
	}

	req, err := httputil.Request(ctx, httputil.RequestOpts{
		URL:    url.String(),
		Method: http.MethodGet,
		Headers: []httputil.Multimap{
			httputil.BasicAuth(azuredevops.UserAgent, cli.Token),
		},
	})
	if err != nil {
		return nil, err
	}

	apiErr := &APIError{}
	val := &TeamProject{}
	err = httputil.Call(cli.HTTPClient, httputil.CallOpts{
		Request: req,
		Handler: httputil.ToJSON(val),
		Validators: []httputil.ResponseHandler{
			httputil.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	return val, err
}
