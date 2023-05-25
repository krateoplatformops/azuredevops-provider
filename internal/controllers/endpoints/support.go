package endpoints

import (
	"context"
	"fmt"

	endpointsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/endpoints/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/endpoints"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/projects"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/pkg/errors"
)

type ProjectReference struct {
	Id           string
	Name         string
	Organization string
}

func (e *external) resolveProjectRef(ctx context.Context, cr *endpointsv1alpha1.Endpoint) (ProjectReference, error) {
	ref := ProjectReference{
		Organization: helpers.StringOrDefault(cr.Spec.Organization, ""),
		Name:         helpers.StringOrDefault(cr.Spec.Project, ""),
	}

	if cr.Spec.PojectRef != nil {
		prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PojectRef)
		if err != nil {
			return ref, errors.Wrapf(err, "unble to resolve TeamProject: %s", cr.Spec.PojectRef.Name)
		}
		if prj != nil {
			ref.Id = prj.Status.Id
			ref.Name = prj.Spec.Name
			ref.Organization = prj.Spec.Organization
		}
	}

	if len(ref.Name) == 0 {
		return ref, fmt.Errorf("missing Project name")
	}

	if len(ref.Organization) == 0 {
		return ref, fmt.Errorf("missing Organization name")
	}

	if len(ref.Id) == 0 {
		tmp, err := projects.Find(ctx, e.azCli, projects.FindOptions{
			Organization: ref.Organization,
			Name:         ref.Name,
		})
		if err != nil {
			return ref, err
		}
		if tmp != nil {
			ref.Id = helpers.String(tmp.Id)
		}
	}

	return ref, nil
}

func (e *external) findEndpoint(ctx context.Context, ref *ProjectReference, cr *endpointsv1alpha1.Endpoint) (*endpoints.ServiceEndpoint, error) {
	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	all, err := endpoints.Find(ctx, e.azCli, endpoints.FindOptions{
		Organization:  ref.Organization,
		Project:       ref.Name,
		EndpointNames: []string{name},
	})
	if err != nil {
		return nil, err
	}

	return &all[0], nil
}

func asAzureDevopsServiceEndpoint(ref *ProjectReference, cr *endpointsv1alpha1.Endpoint) *endpoints.ServiceEndpoint {
	if cr.Spec.Name == nil {
		cr.Spec.Name = helpers.StringPtr(cr.GetName())
	}

	res := &endpoints.ServiceEndpoint{
		Name:          cr.Spec.Name,
		Description:   cr.Spec.Description,
		IsShared:      cr.Spec.IsShared,
		Owner:         cr.Spec.Owner,
		Type:          cr.Spec.Type,
		Url:           cr.Spec.Url,
		Authorization: &endpoints.EndpointAuthorization{},
		Data:          map[string]string{},
		ServiceEndpointProjectReferences: []endpoints.ServiceEndpointProjectReference{
			{
				Name:        cr.Spec.Name,
				Description: cr.Spec.Description,
				ProjectReference: &endpoints.ProjectReference{
					Id:   helpers.StringPtr(ref.Id),
					Name: ref.Name,
				},
			},
		},
	}

	if aut := cr.Spec.Authorization; aut != nil {
		if aut.Scheme != nil {
			res.Authorization.Scheme = aut.Scheme
		}
		if aut.Parameters != nil {
			res.Authorization.Parameters = map[string]string{
				"tenantid":                  helpers.String(aut.Parameters.Tenantid),
				"serviceprincipalId":        helpers.String(aut.Parameters.ServiceprincipalId),
				"authenticationType":        helpers.String(aut.Parameters.AuthenticationType),
				"serviceprincipalKey":       helpers.String(aut.Parameters.ServiceprincipalKey),
				"scope":                     helpers.String(aut.Parameters.Scope),
				"serviceAccountCertificate": helpers.String(aut.Parameters.ServiceAccountCertificate),
				"isCreatedFromSecretYaml":   helpers.String(aut.Parameters.IsCreatedFromSecretYaml),
				"apitoken":                  helpers.String(aut.Parameters.Apitoken),
			}
		}
	}

	if dt := cr.Spec.Data; dt != nil {
		res.Data["environment"] = helpers.String(dt.Environment)
		res.Data["scopeLevel"] = helpers.String(dt.ScopeLevel)
		res.Data["subscriptionId"] = helpers.String(dt.SubscriptionId)
		res.Data["subscriptionName"] = helpers.String(dt.SubscriptionName)
		res.Data["creationMode"] = helpers.String(dt.CreationMode)
		res.Data["authorizationType"] = helpers.String(dt.AuthorizationType)
		res.Data["acceptUntrustedCerts"] = helpers.String(dt.AcceptUntrustedCerts)
	}

	for _, el := range cr.Spec.ServiceEndpointProjectReferences {
		if el.ProjectReference == nil {
			continue
		}
		if helpers.String(el.ProjectReference.Id) == ref.Id {
			continue
		}

		spr := endpoints.ServiceEndpointProjectReference{
			Description:      el.Description,
			Name:             el.Name,
			ProjectReference: &endpoints.ProjectReference{},
		}
		if pr := el.ProjectReference; pr != nil {
			spr.ProjectReference.Id = pr.Id
			spr.ProjectReference.Name = pr.Name
		}

		res.ServiceEndpointProjectReferences = append(res.ServiceEndpointProjectReferences, spr)
	}

	return res
}
