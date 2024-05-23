package checkconfiguration

import (
	"context"
	"errors"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type CreatedBy struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
	UniqueName  string `json:"uniqueName"`
	Descriptor  string `json:"descriptor"`
}

type ModifiedBy struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
	UniqueName  string `json:"uniqueName"`
	Descriptor  string `json:"descriptor"`
}

type Type struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Resource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CheckConfiguration struct {
	CreatedBy  CreatedBy   `json:"createdBy"`
	CreatedOn  string      `json:"createdOn"`
	ModifiedBy ModifiedBy  `json:"modifiedBy"`
	ModifiedOn string      `json:"modifiedOn"`
	Timeout    int         `json:"timeout"`
	Links      interface{} `json:"_links"`
	ID         int         `json:"id"`
	Type       Type        `json:"type"`
	URL        string      `json:"url"`
	Resource   Resource    `json:"resource"`
}

type GetOptions struct {
	// Organization Name
	Organization string
	// ProjectID or Project Name
	Project string
	// CheckID
	CheckID string
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.CheckConfigurations
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

// Get Check configuration by Id
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines/checks/configurations/{id}?api-version=7.0-preview.1
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*CheckConfiguration, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines/checks/configurations", opts.CheckID),
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
	val := &CheckConfiguration{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if val != nil && reflect.DeepEqual(*val, CheckConfiguration{}) {
		return nil, err
	}

	return val, err
}

// Options for the List function
type ListOptions struct {
	// Name of the organization
	Organization string
	// Project ID or project name
	Project      string
	ResourceType *string
	ResourceId   *string
}

type ListResponse struct {
	Value []CheckConfiguration
	Count int
}

// List check configuration by project
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines/checks/configurations?api-version=7.0-preview.1
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListResponse, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}

	if opts.ResourceId == nil {
		return nil, errors.New("ResourceId parameter is required")
	}
	if opts.ResourceType == nil {
		return nil, errors.New("ResourceType parameter is required")
	}
	var queryparams []string
	queryparams = append(apiVersionParams, "resourceId", helpers.String(opts.ResourceId))
	queryparams = append(queryparams, "resourceType", helpers.String(opts.ResourceType))
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines/checks/configurations"),
		Params:  queryparams,
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
	val := &ListResponse{
		Value: []CheckConfiguration{},
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
	Type
}

func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*CheckConfiguration, error) {
	list, err := List(ctx, cli, opts.ListOptions)
	if err != nil {
		return nil, err
	}

	for _, check := range list.Value {
		if strings.EqualFold(check.Type.ID, opts.Type.ID) {
			return &check, nil
		}
	}
	return nil, &httplib.StatusError{StatusCode: http.StatusNotFound, Inner: errors.New("check configuration not found")}
}

type Approver struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
}

type ApprovalSettings struct {
	Approvers                 []Approver `json:"approvers"`
	ExecutionOrder            string     `json:"executionOrder"`
	MinRequiredApprovers      int        `json:"minRequiredApprovers"`
	Instructions              string     `json:"instructions"`
	BlockedApprovers          []string   `json:"blockedApprovers"`
	RequesterCannotBeApprover bool       `json:"requesterCannotBeApprover"`
}

type Approval struct {
	Settings ApprovalSettings `json:"settings"`
	Timeout  int              `json:"timeout"`
	Type     Type             `json:"type"`
	Resource Resource         `json:"resource"`
}

type DefinitionRef struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type TaskCheckSettings struct {
	// Inputs in inline JSON format
	Inputs              interface{}   `json:"inputs"`
	LinkedVariableGroup string        `json:"linkedVariableGroup"`
	RetryInterval       int           `json:"retryInterval"`
	DisplayName         string        `json:"displayName"`
	DefinitionRef       DefinitionRef `json:"definitionRef"`
}

type TaskCheck struct {
	Settings TaskCheckSettings `json:"settings"`
	Timeout  int               `json:"timeout"`
	Type     Type              `json:"type"`
	Resource Resource          `json:"resource"`
}

type ExtendsCheckSetting struct {
	RepositoryType string `json:"repositoryType,omitempty"`
	RepositoryName string `json:"repositoryName,omitempty"`
	RepositoryRef  string `json:"repositoryRef,omitempty"`
	TemplatePath   string `json:"templatePath,omitempty"`
}

type ExtendsCheckSettings struct {
	ExtendsChecks []ExtendsCheckSetting `json:"extendsChecks"`
}

type ExtendsCheck struct {
	Settings ExtendsCheckSettings `json:"settings"`
	Type     Type                 `json:"type"`
	Resource Resource             `json:"resource"`
}

type CheckOptions interface {
	Approval | TaskCheck | ExtendsCheck
}

type CreateOptions[T CheckOptions] struct {
	// Organization Name
	Organization string
	// ProjectID or Project Name
	Project string
	// Resource Approval or TaskCheck
	CheckRes T
}

// Add a check configuration
// POST https://dev.azure.com/{organization}/{project}/_apis/pipelines/checks/configurations?api-version=7.0-preview.1
func Create[T CheckOptions](ctx context.Context, cli *azuredevops.Client, opts CreateOptions[T]) (*CheckConfiguration, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}

	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines/checks/configurations"),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.CheckRes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &CheckConfiguration{}
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

type DeleteOptions struct {
	Organization string
	Project      string
	CheckId      string
}

// Delete check configuration by id
// DELETE https://dev.azure.com/{organization}/{project}/_apis/pipelines/checks/configurations/{id}?api-version=7.0-preview.1
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines/checks/configurations", opts.CheckId),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	return httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		Validators: []httplib.HandleResponseFunc{
			httplib.CheckStatus(http.StatusOK, http.StatusNoContent),
		},
	})
}
