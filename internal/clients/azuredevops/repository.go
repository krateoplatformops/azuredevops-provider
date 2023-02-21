package azuredevops

import (
	"context"
	"net/http"
	"path"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type GitRepository struct {
	Id            *string      `json:"id,omitempty"`
	Name          *string      `json:"name,omitempty"`
	Project       *TeamProject `json:"project,omitempty"`
	DefaultBranch *string      `json:"defaultBranch,omitempty"`
	RemoteUrl     *string      `json:"remoteUrl,omitempty"`
	SshUrl        *string      `json:"sshUrl,omitempty"`
	Url           *string      `json:"url,omitempty"`
}

type GetRepositoryOptions struct {
	Organization string
	ProjectId    string
	RepositoryId string
}

// GetRepository retrieve a git repository.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}?api-version=7.0
func (c *Client) GetRepository(ctx context.Context, opts GetRepositoryOptions) (*GitRepository, error) {
	ub := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories", opts.RepositoryId),
		Params:  []string{apiVersionKey, apiVersionVal},
	})

	req, err := httplib.NewGetRequest(ub)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &GitRepository{}

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

type CreateRepositoryOptions struct {
	Organization string
	ProjectId    string
	Name         string
}

// CreateRepository creates a git repository in a team project.
// POST https://dev.azure.com/{organization}/{project}/_apis/git/repositories?api-version=7.0
func (c *Client) CreateRepository(ctx context.Context, opts CreateRepositoryOptions) (*GitRepository, error) {
	ub := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories"),
		Params:  []string{apiVersionKey, apiVersionVal},
	})

	req, err := httplib.NewPostRequest(ub, httplib.ToJSON(&GitRepository{
		Name: &opts.Name,
		Project: &TeamProject{
			Id: helpers.StringPtr(opts.ProjectId),
		},
	}))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &GitRepository{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusCreated),
		},
	})
	return val, err
}
