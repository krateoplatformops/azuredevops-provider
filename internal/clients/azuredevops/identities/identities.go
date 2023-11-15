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
	IdentityParams
	// (required) The name of the Azure DevOps organization.
	Organization string
}

type UserType string

const (
	BuildService UserType = "build-service"
	AzureGroup   UserType = "azure-group"
)

func (u *UserType) ResolveIdentityDescriptorFromUserType() (string, error) {
	switch *u {
	case BuildService:
		return "Microsoft.TeamFoundation.ServiceIdentity", nil
	case AzureGroup:
		return "Microsoft.TeamFoundation.Identity", nil
	}
	return " ", errors.Errorf("The specified usertype is not valid")
}

type IdentityParams struct {
	Type    UserType
	Project *teamprojects.TeamProject
	// Name is ignored if Type is build-service
	Name string
}

func (resp *IdentityResponse) IdentityMatch(identity *IdentityParams) (*Identity, error) {
	for _, v := range resp.Value {
		resolvedId, err := identity.Type.ResolveIdentityDescriptorFromUserType()
		if err != nil {
			return nil, err
		}
		if strings.Contains(v.Descriptor, resolvedId) {
			if identity.Type == BuildService && v.ProviderDisplayName == identity.Project.Status.Id {
				return &v, nil
			}

			if identity.Type == AzureGroup && v.ProviderDisplayName == fmt.Sprintf("[%s]\\%s", identity.Project.Spec.Name, identity.Name) {
				return &v, nil
			}
		}
	}
	return nil, errors.Errorf("Identity not found")
}

func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*IdentityResponse, error) {

	var filterValue string
	switch opts.Type {
	case BuildService:
		filterValue = opts.Project.Status.Id
	case AzureGroup:
		filterValue = fmt.Sprint("[", opts.Project.Spec.Name, "]", "\\", opts.Name)
	}

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Vssps),
		Path:    path.Join(opts.Organization, "_apis/identities"),
		Params:  []string{"searchFilter", "General", "filterValue", filterValue, "queryMembership", "None", azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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
