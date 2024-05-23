package variablegroups

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type Variable struct {
	Value      string `json:"value"`
	IsSecret   bool   `json:"isSecret"`
	IsReadOnly bool   `json:"isReadOnly"`
}

type ProjectReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type VariableGroupProjectReference struct {
	Name             string           `json:"name"`
	ProjectReference ProjectReference `json:"projectReference"`
}

type VariableGroupBody struct {
	Variables                      map[string]Variable             `json:"variables"`
	Type                           string                          `json:"type"`
	VariableGroupProjectReferences []VariableGroupProjectReference `json:"variableGroupProjectReferences"`
	Name                           string                          `json:"name"`
	Description                    string                          `json:"description"`
}

type CreatedModifiedBy struct {
	DisplayName *string `json:"displayName"`
	ID          string  `json:"id"`
}

type VariableGroupResponse struct {
	Variables                      map[string]Variable             `json:"variables"`
	ID                             int                             `json:"id"`
	Type                           string                          `json:"type"`
	Name                           string                          `json:"name"`
	Description                    string                          `json:"description"`
	CreatedBy                      CreatedModifiedBy               `json:"createdBy"`
	CreatedOn                      string                          `json:"createdOn"`
	ModifiedBy                     CreatedModifiedBy               `json:"modifiedBy"`
	ModifiedOn                     string                          `json:"modifiedOn"`
	IsShared                       bool                            `json:"isShared"`
	VariableGroupProjectReferences []VariableGroupProjectReference `json:"variableGroupProjectReferences"`
}

type GetOptions struct {
	Description     string
	Name            string
	Organization    string
	VariableGroupId int
	Project         string
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.VariableGroups
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

// Get a variable group.
// GET https://dev.azure.com/{organization}/{project}/_apis/distributedtask/variablegroups/{groupId}?api-version=7.1-preview.2
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*VariableGroupResponse, error) {
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/variablegroups", fmt.Sprintf("%d", opts.VariableGroupId))
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
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
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &VariableGroupResponse{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if reflect.DeepEqual(*val, VariableGroupResponse{}) {
		return nil, err
	}

	return val, err
}

type CreateBody struct {
	Name        string
	Description string
}

type CreateOptions struct {
	Organization  string
	Project       string
	VariableGroup *VariableGroupBody
}

// Add a variable group.
// POST https://dev.azure.com/{organization}/_apis/distributedtask/variablegroups?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*VariableGroupResponse, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/variablegroups")
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.VariableGroup))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &VariableGroupResponse{}
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
	Organization      string
	Project           string
	ContinuationToken *string
}

type ListReturn struct {
	Count             int                     `json:"count"`
	Value             []VariableGroupResponse `json:"value"`
	ContinuationToken *string                 `json:"continuationToken"`
}

// Get variable groups.
// GET https://dev.azure.com/{organization}/{project}/_apis/distributedtask/variablegroups?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListReturn, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}

	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/variablegroups")

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
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
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &ListReturn{
		Value: []VariableGroupResponse{},
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type FindOptions struct {
	ListOptions
	VariableGroupName string
}

// Find a VariableGroup by its name.
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*VariableGroupResponse, error) {
	for {
		groups, err := List(ctx, cli, opts.ListOptions)
		if err != nil {
			return nil, err
		}
		for _, group := range groups.Value {
			fmt.Println(group.Name, opts.VariableGroupName)
			if strings.EqualFold(group.Name, opts.VariableGroupName) {
				return &group, nil
			}
		}
		if groups.ContinuationToken == nil {
			break
		}
		opts.ListOptions.ContinuationToken = groups.ContinuationToken
	}

	return nil, nil
}

type DeleteOptions struct {
	Organization    string
	ProjectID       string
	VariableGroupId int
}

// Delete a variable group
// DELETE https://dev.azure.com/{organization}/_apis/distributedtask/variablegroups/{groupId}?projectIds={projectIds}&api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	fullPath := path.Join(opts.Organization, "_apis/distributedtask/variablegroups", fmt.Sprintf("%d", opts.VariableGroupId))
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
		Params:  apiVersionParams,
	}
	if opts.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	ubo.Params = append(ubo.Params, []string{"projectIds", opts.ProjectID}...)

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
	Organization    string
	Project         string
	VariableGroupId int
	VariableGroup   *VariableGroupBody
}

// Update a variable group.
// PUT https://dev.azure.com/{organization}/_apis/distributedtask/variablegroups/{groupId}?api-version=7.1-preview.2
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*VariableGroupResponse, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	fullPath := path.Join(opts.Organization, opts.Project, "_apis/distributedtask/variablegroups", fmt.Sprintf("%d", opts.VariableGroupId))

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    fullPath,
		Params:  apiVersionParams,
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Put(uri.String(), httplib.ToJSON(opts.VariableGroup))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &VariableGroupResponse{}
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
