package environments

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type EnvironmentResourceType string

const (
	generic        EnvironmentResourceType = "generic"
	kubernetes     EnvironmentResourceType = "kubernetes"
	undefined      EnvironmentResourceType = "undefined"
	virtualMachine EnvironmentResourceType = "virtualMachine"
)

type ProjectReference struct {
	// Gets or sets id of the project.
	Id string `json:"id,omitempty"`
	// Gets or sets name of the project.
	Name string `json:"name,omitempty"`
}

type EnvironmentResourceReference struct {
	Id   *int                     `json:"id,omitempty"`
	Name *string                  `json:"name,omitempty"`
	Tags []string                 `json:"tags,omitempty"`
	Type *EnvironmentResourceType `json:"type,omitempty"`
}

type Environment struct {
	CreatedBy      *azuredevops.IdentityRef       `json:"createdBy,omitempty"`
	CreatedOn      *string                        `json:"createdOn,omitempty"`
	Description    *string                        `json:"description,omitempty"`
	Id             *int                           `json:"id,omitempty"`
	LastModifiedBy *azuredevops.IdentityRef       `json:"lastModifiedBy,omitempty"`
	LastModifiedOn *string                        `json:"lastModifiedOn,omitempty"`
	Name           *string                        `json:"name,omitempty"`
	Project        *ProjectReference              `json:"project,omitempty"`
	Resources      []EnvironmentResourceReference `json:"resources,omitempty"`
}

type GetOptions struct {
	Description   string
	Name          string
	Organization  string
	EnvironmentId int
	Project       string
}

// Get an environment by its ID.
// GET https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments/{environmentId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*Environment, error) {
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/environments/", fmt.Sprintf("%d", opts.EnvironmentId))

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
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
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &Environment{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if val != nil && reflect.DeepEqual(*val, Environment{}) {
		return nil, err
	}

	return val, err
}

type CreateBody struct {
	Name        string
	Description string
}

type CreateOptions struct {
	Organization string
	Project      string
	Environment  *Environment
}

// Create an environment.
// POST https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*Environment, error) {
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/environments")
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.Environment))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &Environment{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusCreated),
		},
	})
	return val, err
}

type ListOptions struct {
	Organization string
	Project      string
	IncludeUrls  bool
}

type ListReturn struct {
	Count int           `json:"count"`
	Value []Environment `json:"value"`
}

// Get all environments.
// GET https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) ([]Environment, error) {
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/environments")

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
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
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &ListReturn{
		Value: []Environment{},
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val.Value, err
}

type FindOptions struct {
	Organization    string
	Project         string
	EnvironmentName string
}

// Find an environment by its name.
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*Environment, error) {
	all, err := List(ctx, cli, ListOptions{
		Organization: opts.Organization,
		Project:      opts.Project,
		IncludeUrls:  true,
	})
	if err != nil {
		return nil, err
	}

	for _, el := range all {
		if helpers.String(el.Name) == opts.EnvironmentName {
			return &el, nil
		}
	}

	return nil, nil
}

type DeleteOptions struct {
	Organization  string
	Project       string
	EnvironmentId int
}

// Delete an environment by its ID.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments/{environmentId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/environments/", fmt.Sprintf("%d", opts.EnvironmentId))

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusNoContent),
		},
	})

	return err
}

type UpdateOptions struct {
	Organization  string
	Project       string
	EnvironmentId int
	Environment   *Environment
}

// Update an environment by its ID.
// PATCH https://dev.azure.com/{organization}/{project}/_apis/distributedtask/environments/{environmentId}?api-version=7.0
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*Environment, error) {
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/environments/", fmt.Sprintf("%d", opts.EnvironmentId))

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Patch(uri.String(), httplib.ToJSON(opts.Environment))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &Environment{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusCreated),
		},
	})
	return val, err
}
