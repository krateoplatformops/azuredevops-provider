package memberships

import (
	"context"
	"net/http"
	"path"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
)

// MembershipStateResponse is the response from the Azure DevOps API
// GET https://vssps.dev.azure.com/{organization}/_apis/graph/membershipstates/{subjectDescriptor}?api-version=7.0-preview.1
type MembershipStateResponse struct {
	Links  interface{} `json:"_links"`
	Active bool        `json:"active"`
}

type GetOptions struct {
	Organization      string
	SubjectDescriptor string
}

func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*MembershipStateResponse, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/membershipstates", opts.SubjectDescriptor),
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

	res := &MembershipStateResponse{}
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

type CheckMembershipOptions struct {
	Organization        string
	SubjectDescriptor   string
	ContainerDescriptor string
}

// CheckMembership checks if the user is a member of the group.
// GET https://vssps.dev.azure.com/{organization}/_apis/graph/memberships/{subjectDescriptor}/{containerDescriptor}?api-version=7.0-preview.1
func CheckMembership(ctx context.Context, cli *azuredevops.Client, opts CheckMembershipOptions) error {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/memberships", opts.SubjectDescriptor, opts.ContainerDescriptor),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return err
	}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusOK),
		},
	})
	return err
}

// Create a new membership between a container and subject.
// PUT https://vssps.dev.azure.com/{organization}/_apis/graph/memberships/{subjectDescriptor}/{containerDescriptor}?api-version=7.0-preview.1
func Create(ctx context.Context, cli *azuredevops.Client, opts CheckMembershipOptions) error {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/graph/memberships", opts.SubjectDescriptor, opts.ContainerDescriptor),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal + azuredevops.ApiPreviewFlag + ".1"},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Put(uri.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(&azuredevops.APIError{}, http.StatusOK),
		},
	})
	return err
}
