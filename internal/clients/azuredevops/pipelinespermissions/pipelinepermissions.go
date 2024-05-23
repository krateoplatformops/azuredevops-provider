package pipelinespermissions

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

type Permission struct {
	Authorized   bool                     `json:"authorized,omitempty"`
	AuthorizedBy *azuredevops.IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *azuredevops.Time        `json:"authorizedOn,omitempty"`
}

type PipelinePermission struct {
	Authorized   bool                     `json:"authorized,omitempty"`
	AuthorizedBy *azuredevops.IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *azuredevops.Time        `json:"authorizedOn,omitempty"`
	Id           interface{}              `json:"id,omitempty"`
}

func (p *PipelinePermission) GetId() string {
	return fmt.Sprintf("%v", p.Id)
}

type ResourcePipelinePermissions struct {
	AllPipelines *Permission           `json:"allPipelines,omitempty"`
	Pipelines    []PipelinePermission  `json:"pipelines,omitempty"`
	Resource     *azuredevops.Resource `json:"resource,omitempty"`
}

// Options for the Get Pipeline Permissions ForResource function
type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	Project string
	// (required)
	ResourceType string // [repository, endpoint, variablegroup, environment, queue]
	// (required)
	ResourceId string
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.PipelinePermissions
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

// [Preview API] Given a ResourceType and ResourceId, returns authorized definitions for that resource.
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines/pipelinepermissions/{resourceType}/{resourceId}?api-version=7.0-preview.1
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*ResourcePipelinePermissions, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines/pipelinepermissions", opts.ResourceType, opts.ResourceId),
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
	val := &ResourcePipelinePermissions{
		AllPipelines: &Permission{},
		Pipelines:    []PipelinePermission{},
		Resource:     &azuredevops.Resource{},
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if val != nil && reflect.DeepEqual(*val, ResourcePipelinePermissions{
		AllPipelines: &Permission{},
		Resource:     &azuredevops.Resource{},
	}) {
		return nil, err
	}

	return val, err
}

// Options for the Update Pipeline Permissions ForResource function
type UpdateOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	Project string
	// (required)
	ResourceType string // [repository, endpoint, variablegroup, environment, queue]
	// (required)
	ResourceId string

	ResourceAuthorization *ResourcePipelinePermissions
}

// Authorizes/Unauthorizes a list of definitions for a given resource.
// PATCH https://dev.azure.com/{organization}/{project}/_apis/pipelines/pipelinepermissions/{resourceType}/{resourceId}?api-version=7.0-preview.1
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*ResourcePipelinePermissions, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag}
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines/pipelinepermissions", opts.ResourceType, opts.ResourceId),
		Params:  apiVersionParams,
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Patch(uri.String(), httplib.ToJSON(opts.ResourceAuthorization))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &ResourcePipelinePermissions{
		AllPipelines: &Permission{},
		Pipelines:    []PipelinePermission{},
		Resource:     &azuredevops.Resource{},
	}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusAccepted),
		},
	})
	return val, err
}
