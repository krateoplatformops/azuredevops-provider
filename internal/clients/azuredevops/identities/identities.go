package identities

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	teamprojects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
	"github.com/pkg/errors"
)

type IdentityResponse struct {
	Count int        `json:"count"`
	Value []Identity `json:"value"`
}
type Identity struct {
	ID                  string   `json:"id"`
	Descriptor          string   `json:"descriptor"`
	SubjectDescriptor   string   `json:"subjectDescriptor"`
	ProviderDisplayName string   `json:"providerDisplayName"`
	CustomDisplayName   string   `json:"customDisplayName"`
	IsActive            bool     `json:"isActive"`
	Members             []string `json:"members"`
	MemberOf            []string `json:"memberOf"`
	MemberIds           []string `json:"memberIds"`
	ResourceVersion     int      `json:"resourceVersion"`
	MetaTypeID          int      `json:"metaTypeId"`
}

type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	ProjectID string
}

type UserType string

const (
	BuildService UserType = "build-service"
)

func (u *UserType) ResolveIdentityDescriptorFromUserType() (string, error) {
	switch *u {
	case BuildService:
		return "Microsoft.TeamFoundation.ServiceIdentity", nil
	}
	return " ", errors.Errorf("The specified usertype is not valid")
}

func (resp *IdentityResponse) IdentityMatch(userType UserType, proj *teamprojects.TeamProject) (*Identity, error) {
	for _, v := range resp.Value {
		resolvedId, err := userType.ResolveIdentityDescriptorFromUserType()
		if err != nil {
			return nil, err
		}
		fmt.Println("ProviderName:", v.ProviderDisplayName)
		fmt.Println("Pj id:", proj.Status.Id)
		if strings.Contains(v.Descriptor, resolvedId) {

			if userType == BuildService && v.ProviderDisplayName == proj.Status.Id {
				return &v, nil
			}
		}
	}
	return nil, errors.Errorf("Identity not found")
}

func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*IdentityResponse, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/identities"),
		Params:  []string{"searchFilter", "General", "filterValue", opts.ProjectID, azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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
	val := &IdentityResponse{
		Value: []Identity{},
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
