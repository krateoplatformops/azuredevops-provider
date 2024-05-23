package teams

import (
	"context"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type TeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	IdentityURL string `json:"identityUrl"`
	ProjectName string `json:"projectName"`
	ProjectID   string `json:"projectId"`
}

type TeamListResponse struct {
	Count int            `json:"count"`
	Value []TeamResponse `json:"value"`
}

type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// Project ID
	ProjectID string
	// Team ID
	TeamID string
}

type ListOptions struct {
	Organization string
	ProjectID    string
}

type TeamData struct {
	Name        string
	Description *string
}

type CreateOptions struct {
	Organization string
	ProjectID    string
	TeamData
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.Teams
		if apiVersion != nil {
			if strings.EqualFold(*apiVersion, "none") {
				apiVersionParams = nil
				isNone = true
			} else {
				apiVersionParams = []string{azuredevops.ApiVersionKey, helpers.String(apiVersion)}
			}
		}
	}
	return apiVersionParams, isNone
}

// Get a specific team.
// GET https://dev.azure.com/{organization}/_apis/projects/{projectId}/teams/{teamId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*TeamResponse, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectID, "teams", opts.TeamID),
		Params:  apiVersionParams,
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}

	res := &TeamResponse{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(res),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusOK),
		},
	})
	if reflect.DeepEqual(*res, TeamResponse{}) {
		return nil, err
	}

	return res, err
}

// Get a list of teams.
// GET https://dev.azure.com/{organization}/_apis/projects/{projectId}/teams?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*TeamListResponse, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectID, "teams"),
		Params:  apiVersionParams,
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}

	res := &TeamListResponse{
		Value: []TeamResponse{},
	}
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

type FindTeamByNameOptions struct {
	ListOptions
	TeamName  string
	ProjectID string
}

func FindTeamByName(ctx context.Context, cli *azuredevops.Client, opts FindTeamByNameOptions) (*TeamResponse, error) {
	teams, err := List(ctx, cli, opts.ListOptions)
	if err != nil {
		return nil, err
	}
	for _, team := range teams.Value {
		if strings.EqualFold(team.Name, opts.TeamName) && strings.EqualFold(team.ProjectID, opts.ProjectID) {
			return &team, nil
		}
	}
	return nil, nil
}

// Create a team in a team project.
// POST https://dev.azure.com/{organization}/_apis/projects/{projectId}/teams?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*TeamResponse, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectID, "teams"),
		Params:  apiVersionParams,
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.TeamData))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	res := &TeamResponse{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(res),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusCreated),
		},
	})

	return res, err
}

type DeleteOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// Project ID
	ProjectID string
	// Team ID
	TeamID string
}

// Delete a team.
// DELETE https://dev.azure.com/{organization}/_apis/projects/{projectId}/teams/{teamId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectID, "teams", opts.TeamID),
		Params:  apiVersionParams,
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusNoContent),
		},
	})

	return err
}
