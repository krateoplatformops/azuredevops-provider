package users

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

type GetOptions struct {
	Organization   string
	UserDescriptor string
}

type UserResource struct {
	SubjectKind   string      `json:"subjectKind"`
	Domain        string      `json:"domain"`
	PrincipalName string      `json:"principalName"`
	MailAddress   string      `json:"mailAddress"`
	Origin        string      `json:"origin"`
	OriginID      *string     `json:"originId"`
	DisplayName   string      `json:"displayName"`
	Links         interface{} `json:"_links"`
	URL           string      `json:"url"`
	Descriptor    string      `json:"descriptor"`
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.Users
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

// Get retrieves information about a user.
// GET https://vssps.dev.azure.com/{organization}/_apis/graph/users/{userDescriptor}?api-version=7.0-preview.1
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*UserResource, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/users", opts.UserDescriptor),
		Params:  apiVersionParams,
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
	val := &UserResource{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if err != nil {
		return nil, err
	}
	if val != nil && reflect.DeepEqual(*val, UserResource{}) {
		return nil, err
	}

	return val, nil
}

type ListOptions struct {
	Organization      string
	ContinuationToken *string
}

type ListResponse struct {
	Count             int            `json:"count"`
	Value             []UserResource `json:"value"`
	ContinuationToken *string        `json:"continuationToken"`
}

// GET https://vssps.dev.azure.com/{organization}/_apis/graph/users?api-version=7.0-preview.1
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListResponse, error) {

	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}

	var queryparams []string
	queryparams = append(queryparams, apiVersionParams...)
	if opts.ContinuationToken != nil {
		queryparams = append(queryparams, "continuationToken", helpers.String(opts.ContinuationToken))
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/users"),
		Params:  queryparams,
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
	val := &ListResponse{
		Value: []UserResource{},
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if err != nil {
		return nil, err
	}
	if val != nil && reflect.DeepEqual(*val, UserResource{}) {
		return nil, err
	}

	return val, nil
}

type FindUserByNameOptions struct {
	ListOptions
	PrincipalName string
}

func FindUserByName(ctx context.Context, cli *azuredevops.Client, opts FindUserByNameOptions) (*UserResource, error) {
	var continuationToken *string
	opts.ListOptions.ContinuationToken = continuationToken
	for {
		users, err := List(ctx, cli, opts.ListOptions)
		if err != nil {
			return nil, err
		}
		for _, user := range users.Value {
			if strings.EqualFold(user.PrincipalName, opts.PrincipalName) || strings.EqualFold(user.DisplayName, opts.PrincipalName) {
				return &user, nil
			}
		}
		if users.ContinuationToken == nil {
			break
		}
		opts.ListOptions.ContinuationToken = users.ContinuationToken
	}
	return nil, nil
}

type Identifiers interface {
	PrincipalName | OriginID
}

type PrincipalName struct {
	PrincipalName string `json:"principalName"`
}
type OriginID struct {
	OriginID string `json:"originId"`
}

type CreateOptions[T Identifiers] struct {
	Organization     string
	GroupDescriptors []string
	Identifier       T
}

// Options for the Create User function
// POST https://vssps.dev.azure.com/{organization}/_apis/graph/users?api-version=7.0-preview.1
func Create[T Identifiers](ctx context.Context, cli *azuredevops.Client, opts CreateOptions[T]) (*UserResource, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}
	var queryParams []string
	queryParams = append(queryParams, apiVersionParams...)

	// Contains the list of group descriptors to add the user to separeted by comma
	var groupsString string
	for _, group := range opts.GroupDescriptors {
		groupsString += group + ","
	}
	queryParams = append(queryParams, "groupDescriptors", groupsString)

	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/users"),
		Params:  queryParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.Identifier))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &UserResource{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusCreated),
		},
	})
	if err != nil {
		return nil, err
	}
	if val != nil && reflect.DeepEqual(*val, UserResource{}) {
		return nil, err
	}

	return val, nil
}

type DeleteOptions struct {
	Organization   string
	UserDescriptor string
}

// DELETE https://vssps.dev.azure.com/{organization}/_apis/graph/users/{userDescriptor}?api-version=7.0-preview.1
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/users", opts.UserDescriptor),
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

	apiErr := &azuredevops.APIError{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusNoContent),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
