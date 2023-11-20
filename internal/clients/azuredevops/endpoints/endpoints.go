package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
)

// Represents the authorization used for service endpoint.
type EndpointAuthorization struct {
	// Gets or sets the parameters for the selected authorization scheme.
	Parameters map[string]string `json:"parameters,omitempty"`
	// Gets or sets the scheme used for service endpoint authentication.
	Scheme *string `json:"scheme,omitempty"`
}

type ProjectReference struct {
	Id   *string `json:"id,omitempty"`
	Name string  `json:"name,omitempty"`
}

type ServiceEndpointProjectReference struct {
	// Gets or sets description of the service endpoint.
	Description *string `json:"description,omitempty"`
	// Gets or sets name of the service endpoint.
	Name *string `json:"name,omitempty"`
	// Gets or sets project reference of the service endpoint.
	ProjectReference *ProjectReference `json:"projectReference,omitempty"`
}

// Represents an endpoint which may be used by an orchestration job.
type ServiceEndpoint struct {
	// This is a deprecated field.
	AdministratorsGroup *azuredevops.IdentityRef `json:"administratorsGroup,omitempty"`
	// Gets or sets the authorization data for talking to the endpoint.
	Authorization *EndpointAuthorization `json:"authorization,omitempty"`
	// Gets or sets the identity reference for the user who created the Service endpoint.
	CreatedBy *azuredevops.IdentityRef `json:"createdBy,omitempty"`
	Data      map[string]string        `json:"data,omitempty"`
	// Gets or sets the description of endpoint.
	Description *string `json:"description,omitempty"`
	// This is a deprecated field.
	GroupScopeId *string `json:"groupScopeId,omitempty"`
	// Gets or sets the identifier of this endpoint.
	Id *string `json:"id,omitempty"`
	// EndPoint state indicator
	IsReady *bool `json:"isReady,omitempty"`
	// Indicates whether service endpoint is shared with other projects or not.
	IsShared *bool `json:"isShared,omitempty"`
	// Gets or sets the friendly name of the endpoint.
	Name *string `json:"name,omitempty"`
	// Error message during creation/deletion of endpoint
	OperationStatus interface{} `json:"operationStatus,omitempty"`
	// Owner of the endpoint Supported values are "library", "agentcloud"
	Owner *string `json:"owner,omitempty"`
	// Gets or sets the identity reference for the readers group of the service endpoint.
	ReadersGroup *azuredevops.IdentityRef `json:"readersGroup,omitempty"`
	// All other project references where the service endpoint is shared.
	ServiceEndpointProjectReferences []ServiceEndpointProjectReference `json:"serviceEndpointProjectReferences,omitempty"`
	// Gets or sets the type of the endpoint.
	Type *string `json:"type,omitempty"`
	// Gets or sets the url of the endpoint.
	Url *string `json:"url,omitempty"`
}

type FindOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) Project ID or project name
	Project string
	// (required) Names of the service endpoints.
	EndpointNames []string
	// (optional) Type of the service endpoints.
	Type string
	// (optional) Authorization schemes used for service endpoints.
	AuthSchemes []string
	// (optional) Owner for service endpoints.
	Owner string
	// (optional) Failed flag for service endpoints.
	IncludeFailed bool
	// (optional) Flag to include more details for service endpoints. This is for internal use only and the flag will be treated as false for all other requests
	//IncludeDetails *bool
}

type FindResult struct {
	Count  int               `json:"count"`
	Values []ServiceEndpoint `json:"value,omitempty"`
}

// GET https://dev.azure.com/{organization}/{project}/_apis/serviceendpoint/endpoints?endpointNames={endpointNames}&type={type}&authSchemes={authSchemes}&owner={owner}&includeFailed={includeFailed}&includeDetails={includeDetails}&api-version=7.0
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) ([]ServiceEndpoint, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	if len(opts.EndpointNames) > 0 {
		params = append(params, "endpointNames", strings.Join(opts.EndpointNames, ","))
	}
	if len(opts.Type) > 0 {
		params = append(params, "type", opts.Type)
	}
	if len(opts.AuthSchemes) > 0 {
		params = append(params, "authSchemes", strings.Join(opts.AuthSchemes, ","))
	}
	if len(opts.Owner) > 0 {
		params = append(params, "owner", opts.Owner)
	}
	params = append(params, "includeFailed", fmt.Sprintf("%t", opts.IncludeFailed))

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
			Path:    path.Join(opts.Organization, opts.Project, "_apis/serviceendpoint/endpoints"),
			Params:  params,
		}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := []ServiceEndpoint{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		ResponseHandler: func(res *http.Response) error {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			all := &FindResult{}
			if err = json.Unmarshal(data, &all); err != nil {
				return err
			}

			val = append(val, all.Values...)

			return nil
		},

		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if len(val) == 0 {
		return nil, &httplib.StatusError{
			StatusCode: http.StatusNotFound,
			Inner:      fmt.Errorf("endpoint(s) [%s] not found", strings.Join(opts.EndpointNames, ",")),
		}
	}

	return val, err
}

// GET https://dev.azure.com/{organization}/{project}/_apis/serviceendpoint/endpoints?type={type}&authSchemes={authSchemes}&endpointIds={endpointIds}&owner={owner}&includeFailed={includeFailed}&includeDetails={includeDetails}&api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) ([]ServiceEndpoint, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	if len(opts.EndpointIds) > 0 {
		params = append(params, "endpointIds", strings.Join(opts.EndpointIds, ","))
	}
	if len(opts.Type) > 0 {
		params = append(params, "type", opts.Type)
	}
	if len(opts.AuthSchemes) > 0 {
		params = append(params, "authSchemes", strings.Join(opts.AuthSchemes, ","))
	}
	if len(opts.Owner) > 0 {
		params = append(params, "owner", opts.Owner)
	}
	params = append(params, "includeFailed", fmt.Sprintf("%t", opts.IncludeFailed))

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
			Path:    path.Join(opts.Organization, "_apis/serviceendpoint/endpoints"),
			Params:  params,
		}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := []ServiceEndpoint{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		ResponseHandler: func(res *http.Response) error {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			all := &FindResult{}
			if err = json.Unmarshal(data, &all); err != nil {
				return err
			}

			val = append(val, all.Values...)

			return nil
		},

		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if len(val) == 0 {
		return nil, &httplib.StatusError{
			StatusCode: http.StatusNotFound,
			Inner:      fmt.Errorf("no results"),
		}
	}

	return val, err
}

type ListOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) Project ID or project name
	Project string
	// (optional) Type of the service endpoints.
	Type string
	// (optional) Authorization schemes used for service endpoints.
	AuthSchemes []string
	// (optional) Ids of the service endpoints.
	EndpointIds []string
	// (optional) Owner for service endpoints.
	Owner string
	// (optional) Failed flag for service endpoints.
	IncludeFailed bool
	// (optional) Flag to include more details for service endpoints. This is for internal use only and the flag will be treated as false for all other requests
	//IncludeDetails *bool
}

type CreateOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) Service endpoint to create
	Endpoint *ServiceEndpoint
}

// POST https://dev.azure.com/{organization}/_apis/serviceendpoint/endpoints?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*ServiceEndpoint, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/serviceendpoint/endpoints"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.Endpoint))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &ServiceEndpoint{}

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

type DeleteOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) project Ids from which endpoint needs to be deleted
	ProjectIds []string
	// (required) The agent queue to remove
	EndpointId string
	// (optional) delete the spn created by endpoint
	Deep *bool
}

// DELETE https://dev.azure.com/{organization}/_apis/serviceendpoint/endpoints/{endpointId}?projectIds={projectIds}&deep={deep}&api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	if len(opts.ProjectIds) == 0 {
		return fmt.Errorf("ProjectIds slice cannot be emtpy")
	}

	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	params = append(params, "projectIds", strings.Join(opts.ProjectIds, ","))

	if opts.Deep != nil {
		params = append(params, "deep", fmt.Sprintf("%t", *opts.Deep))
	}

	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/serviceendpoint/endpoints/", opts.EndpointId),
		Params:  params,
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

type GetOptions struct {
	// (required) Name of the organization
	Organization string
	// (optional) Project ID or project name
	Project string
	// (required) The agent queue to get information about
	EndpointId string
}

// GET https://dev.azure.com/{organization}/{project}/_apis/serviceendpoint/endpoints/{endpointId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*ServiceEndpoint, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/serviceendpoint/endpoints", opts.EndpointId),
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
	val := &ServiceEndpoint{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if val != nil && reflect.DeepEqual(*val, ServiceEndpoint{}) {
		return nil, err
	}

	return val, err
}
