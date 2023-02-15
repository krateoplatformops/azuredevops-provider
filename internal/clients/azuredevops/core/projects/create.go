package projects

import (
	"context"
	"net/http"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httputil"
)

type CreateProjectOpts struct {
	Organization string
	TeamProject  *TeamProject
}

// Queues a project to be created. Use the GetOperation to periodically check for create project status.
// POST https://dev.azure.com/{organization}/_apis/projects?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateProjectOpts) (*OperationReference, error) {
	url, err := httputil.BuildURL(cli.BaseURL, []string{
		opts.Organization,
		"_apis/projects",
	}, azuredevops.APIVersion)
	if err != nil {
		return nil, err
	}

	req, err := httputil.Request(ctx, httputil.RequestOpts{
		URL:    url.String(),
		Method: http.MethodPost,
		Headers: []httputil.Multimap{
			httputil.BasicAuth(azuredevops.UserAgent, cli.Token),
			httputil.ContentType("application/json"),
		},
		GetBody: httputil.BodyJSON(opts.TeamProject),
	})
	if err != nil {
		return nil, err
	}

	apiErr := &APIError{}
	val := &OperationReference{}
	err = httputil.Call(cli.HTTPClient, httputil.CallOpts{
		Request: req,
		Handler: httputil.ToJSON(val),
		Validators: []httputil.ResponseHandler{
			httputil.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	return val, err
}
