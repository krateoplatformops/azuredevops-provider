package endpoints

import (
	"context"
	"fmt"

	endpointsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/endpoints/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/endpoints"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/projects"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	if cr.Spec.ProjectRef != nil {
		prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
		if err != nil {
			return ref, errors.Wrapf(err, "unable to resolve TeamProject: %s", cr.Spec.ProjectRef.Name)
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

func asAzureDevopsServiceEndpoint(ctx context.Context, kube client.Client, ref *ProjectReference, cr *endpointsv1alpha1.Endpoint, originEndpoint endpoints.ServiceEndpoint) (*endpoints.ServiceEndpoint, error) {
	if cr.Spec.Name == nil {
		cr.Spec.Name = helpers.StringPtr(cr.GetName())
	}

	res := originEndpoint
	res.Name = helpers.StringPtr(helpers.StringOrDefault(cr.Spec.Name, helpers.String(res.Name)))
	res.Description = helpers.StringPtr(helpers.StringOrDefault(cr.Spec.Description, helpers.String(res.Description)))
	res.IsShared = helpers.BoolPtr(helpers.BoolOrDefault(cr.Spec.IsShared, helpers.Bool(res.IsShared)))
	res.Owner = helpers.StringPtr(helpers.StringOrDefault(cr.Spec.Owner, helpers.String(res.Owner)))
	res.Type = helpers.StringPtr(helpers.StringOrDefault(cr.Spec.Type, helpers.String(res.Type)))
	res.Url = helpers.StringPtr(helpers.StringOrDefault(cr.Spec.Url, helpers.String(res.Url)))
	res.Authorization = &endpoints.EndpointAuthorization{}
	if res.Data == nil {
		res.Data = map[string]string{}
	}
	if res.ServiceEndpointProjectReferences == nil {
		res.ServiceEndpointProjectReferences = []endpoints.ServiceEndpointProjectReference{}
	}
	projRef := endpoints.ServiceEndpointProjectReference{
		Name:        cr.Spec.Name,
		Description: cr.Spec.Description,
		ProjectReference: &endpoints.ProjectReference{
			Id:   helpers.StringPtr(ref.Id),
			Name: ref.Name,
		},
	}
	if !containsRef(res.ServiceEndpointProjectReferences, projRef) {
		res.ServiceEndpointProjectReferences = append(res.ServiceEndpointProjectReferences, projRef)
	}

	if aut := cr.Spec.Authorization; aut != nil {
		if aut.Scheme != nil {
			res.Authorization.Scheme = aut.Scheme
		}
		if aut.Parameters != nil {
			res.Authorization.Parameters = map[string]string{}
			addEventually(res.Authorization.Parameters, "tenantid", aut.Parameters.Tenantid)
			addEventually(res.Authorization.Parameters, "serviceprincipalId", aut.Parameters.ServiceprincipalId)
			addEventually(res.Authorization.Parameters, "authenticationType", aut.Parameters.AuthenticationType)
			addEventually(res.Authorization.Parameters, "serviceprincipalKey", aut.Parameters.ServiceprincipalKey)
			addEventually(res.Authorization.Parameters, "scope", aut.Parameters.Scope)
			addEventually(res.Authorization.Parameters, "serviceAccountCertificate", aut.Parameters.ServiceAccountCertificate)
			addEventually(res.Authorization.Parameters, "isCreatedFromSecretYaml", aut.Parameters.IsCreatedFromSecretYaml)
			addEventually(res.Authorization.Parameters, "apitoken", aut.Parameters.Apitoken)
		}
	} else {
		res.Authorization = originEndpoint.Authorization
	}

	if dt := cr.Spec.Data; dt != nil {
		addEventually(res.Data, "environment", dt.Environment)
		addEventually(res.Data, "scopeLevel", dt.ScopeLevel)
		addEventually(res.Data, "subscriptionId", dt.SubscriptionId)
		addEventually(res.Data, "subscriptionName", dt.SubscriptionName)
		addEventually(res.Data, "creationMode", dt.CreationMode)
		addEventually(res.Data, "authorizationType", dt.AuthorizationType)
		addEventually(res.Data, "acceptUntrustedCerts", dt.AcceptUntrustedCerts)
	}

	for _, el := range cr.Spec.ServiceEndpointProjectReferences {
		projectUser, err := resolvers.ResolveTeamProject(ctx, kube, &rtv1.Reference{
			Name:      el.ProjectRef.Name,
			Namespace: el.ProjectRef.Namespace,
		})
		if err != nil {
			return nil, err
		}
		if projectUser == nil {
			return nil, errors.Errorf("Project with name %s and namespace %s not found", el.ProjectRef.Name, el.ProjectRef.Namespace)
		}

		spr := endpoints.ServiceEndpointProjectReference{
			Description:      el.Description,
			Name:             el.Name,
			ProjectReference: &endpoints.ProjectReference{},
		}
		spr.ProjectReference.Id = helpers.StringPtr(projectUser.Status.Id)
		spr.ProjectReference.Name = projectUser.Spec.Name

		if !containsRef(res.ServiceEndpointProjectReferences, spr) {
			res.ServiceEndpointProjectReferences = append(res.ServiceEndpointProjectReferences, spr)
		}
	}

	return &res, nil
}

func addEventually(dict map[string]string, key string, val *string) {
	if val == nil {
		return
	}

	if s := helpers.String(val); len(s) > 0 {
		dict[key] = s
	}
}

func containsRef(a []endpoints.ServiceEndpointProjectReference, b endpoints.ServiceEndpointProjectReference) bool {
	for _, el := range a {
		if helpers.String(el.Name) == helpers.String(b.Name) && el.ProjectReference.Name == b.ProjectReference.Name && helpers.String(el.ProjectReference.Id) == helpers.String(b.ProjectReference.Id) {
			return true
		}
	}
	return false
}
func getRefDiff(a, b []endpoints.ServiceEndpointProjectReference) (diff []endpoints.ServiceEndpointProjectReference) {
	for _, el := range a {
		if !containsRef(b, el) {
			diff = append(diff, el)
		}
	}
	return
}
