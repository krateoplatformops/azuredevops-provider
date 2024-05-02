package groups

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type GroupResponse struct {
	SubjectKind   string      `json:"subjectKind"`
	Description   string      `json:"description"`
	Domain        string      `json:"domain"`
	PrincipalName string      `json:"principalName"`
	MailAddress   string      `json:"mailAddress"`
	Origin        string      `json:"origin"`
	OriginID      string      `json:"originId"`
	DisplayName   string      `json:"displayName"`
	Links         interface{} `json:"_links"`
	URL           string      `json:"url"`
	Descriptor    string      `json:"descriptor"`
}

type GroupListResponse struct {
	Count             int             `json:"count"`
	Value             []GroupResponse `json:"value"`
	ContinuationToken *string         `json:"continuationToken"`
}

// Options for the Get Pipeline Permissions ForResource function
type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	GroupDescriptor string
}

type ListOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization      string
	ContinuationToken *string
}

type GroupDescription struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type SetGroupMembership struct {
	OriginID string `json:"originId"`
}

type GroupData interface {
	GroupDescription | SetGroupMembership
}
type CreateOptions[T GroupData] struct {
	Organization    string  `json:"organization"`
	GroupData       T       `json:"groupData"`
	ScopeDescriptor *string `json:"scopeDescriptor"`
	//Comma separated list of group descriptors
	GroupDescriptors []string `json:"groupDescriptors"`
}

// Get a group by its descriptor.
// The group will be returned even if it has been deleted from the account or has had all its memberships deleted.
// https://vssps.dev.azure.com/{organization}/_apis/graph/groups/{groupDescriptor}?api-version=7.0-preview.1
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*GroupResponse, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/groups", opts.GroupDescriptor),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}

	res := &GroupResponse{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(res),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusOK),
		},
	})
	if reflect.DeepEqual(*res, GroupResponse{}) {
		return nil, err
	}

	return res, err
}

// Get a list of all groups in the current scope (usually organization or account).
// https://vssps.dev.azure.com/{organization}/_apis/graph/groups?api-version=7.0-preview.1
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*GroupListResponse, error) {
	queryparams := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	if opts.ContinuationToken != nil {
		queryparams = append(queryparams, "continuationToken", helpers.String(opts.ContinuationToken))
	}
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/groups"),
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

	res := &GroupListResponse{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: FromJSON(res),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusOK),
		},
	})

	return res, err
}

type FindGroupByNameOptions struct {
	ListOptions
	GroupName string
	ProjectID *string
}

func FindGroupByName(ctx context.Context, cli *azuredevops.Client, opts FindGroupByNameOptions) (*GroupResponse, error) {
	var continuationToken *string
	opts.ListOptions.ContinuationToken = continuationToken
	for {
		groups, err := List(ctx, cli, opts.ListOptions)
		if err != nil {
			return nil, err
		}
		for _, group := range groups.Value {
			domain := path.Base(group.Domain)
			if strings.EqualFold(group.DisplayName, opts.GroupName) &&
				(opts.ProjectID == nil || //if projectID is not provided, return the first group with the name - organization is from the api endpoint
					strings.EqualFold(domain, helpers.String(opts.ProjectID))) {
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

// Create a new Azure DevOps group.
// POST https://vssps.dev.azure.com/{organization}/_apis/graph/groups?scopeDescriptor={scopeDescriptor}&groupDescriptors={groupDescriptors}&api-version=7.0-preview.1
func Create[T GroupData](ctx context.Context, cli *azuredevops.Client, opts CreateOptions[T]) (*GroupResponse, error) {
	queryParams := []string{}
	if opts.ScopeDescriptor != nil {
		queryParams = append(queryParams, "scopeDescriptor", *opts.ScopeDescriptor)
	}
	if len(opts.GroupDescriptors) > 0 {
		queryParams = append(queryParams, "groupDescriptors", strings.Join(opts.GroupDescriptors, ","))
	}

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/groups"),
		Params:  append(queryParams, azuredevops.ApiVersionKey, azuredevops.ApiVersionVal+azuredevops.ApiPreviewFlag+".1"),
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.GroupData))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	res := &GroupResponse{}
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
	// (required) Project ID or project name
	GroupDescriptor string
}

// Delete a group.
// DELETE https://vssps.dev.azure.com/{organization}/_apis/graph/groups/{groupDescriptor}?api-version=6.1-preview.1
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/groups", opts.GroupDescriptor),
		Params:  []string{azuredevops.ApiVersionKey, "6.1" + azuredevops.ApiPreviewFlag + ".1"},
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

// FromJSON decodes a response as a JSON object.
func FromJSON(v *GroupListResponse) httplib.HandleResponseFunc {
	return func(res *http.Response) error {
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		header := res.Header.Get("X-Ms-Continuationtoken")
		if header != "" {
			v.ContinuationToken = helpers.StringPtr(header)
		}
		if err = json.Unmarshal(data, v); err != nil {
			return err
		}
		return nil
	}
}
