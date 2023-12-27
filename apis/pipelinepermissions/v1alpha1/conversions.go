package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha2"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	"github.com/krateoplatformops/provider-runtime/pkg/errors"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ConvertTo converts this PipelinePermission to the Hub version.
func (src *PipelinePermission) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.PipelinePermission)
	cli, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if err != nil {
		return err
	}
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.ManagedSpec = src.Spec.ManagedSpec
	dst.Spec.ConnectorConfigRef = src.Spec.ConnectorConfigRef
	dst.Spec.AuthorizeAll = src.Spec.Authorize
	teamproject, err := resolvers.FindTeamProjectRef(context.TODO(), cli, src.Spec.Project)
	if err != nil {
		return err
	}
	dst.Spec.ProjectRef = &rtv1.Reference{
		Name:      teamproject.Name,
		Namespace: teamproject.Namespace,
	}
	dst.Spec.Resource = &v1alpha2.Resource{}
	dst.Spec.Resource.Type = src.Spec.Resource.Type
	id := helpers.String(src.Spec.Resource.Id)
	ty := helpers.String(dst.Spec.Resource.Type)
	finder := resolvers.GetFinderFromType(ty)
	if ty == string(v1alpha2.GitRepository) {
		arr := strings.Split(id, ".")
		if len(arr) > 1 {
			id = arr[1]
		}
	}
	ref, err := finder(context.TODO(), cli, id)
	if err != nil {
		return err
	}
	dst.Spec.Resource.ResourceRef = &rtv1.Reference{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}

	return nil
}

// ConvertFrom converts from the Hub version to this version.
func (dst *PipelinePermission) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.PipelinePermission)
	ctx := context.TODO()
	cli, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.ManagedSpec = src.Spec.ManagedSpec
	dst.Spec.ConnectorConfigRef = src.Spec.ConnectorConfigRef
	project, err := resolvers.ResolveTeamProject(ctx, cli, src.Spec.ProjectRef)
	if err != nil {
		return err
	}
	dst.Spec.Project = helpers.StringOrDefault(helpers.StringPtr(project.Status.Id), project.Spec.Name)
	dst.Spec.Organization = helpers.String(&project.Spec.Organization)
	dst.Spec.Authorize = src.Spec.AuthorizeAll

	dst.Spec.Resource = &Resource{}
	dst.Spec.Resource.Type = src.Spec.Resource.Type
	ty := helpers.String(dst.Spec.Resource.Type)
	var id, name string
	switch ty {
	case string(v1alpha2.TeamProject):
		res, _ := resolvers.ResolveTeamProject(ctx, cli, src.Spec.Resource.ResourceRef)
		id = res.Status.Id
		name = res.Spec.Name
	case string(v1alpha2.GitRepository):
		res, _ := resolvers.ResolveGitRepository(ctx, cli, src.Spec.Resource.ResourceRef)
		id = res.Status.Id
		name = res.Spec.Name
	case string(v1alpha2.Queue):
		res, _ := resolvers.ResolveQueue(ctx, cli, src.Spec.Resource.ResourceRef)
		id = fmt.Sprintf("%v", helpers.Int(res.Status.Id))
		name = helpers.String(res.Spec.Name)
	case string(v1alpha2.Environment):
		res, _ := resolvers.ResolveEnvironment(ctx, cli, src.Spec.Resource.ResourceRef)
		id = fmt.Sprintf("%v", helpers.Int(res.Status.Id))
		name = helpers.String(res.Spec.Name)
	case string(v1alpha2.Endpoint):
		res, _ := resolvers.ResolveEndpoint(ctx, cli, src.Spec.Resource.ResourceRef)
		id = helpers.String(res.Status.Id)
		name = helpers.String(res.Spec.Name)
	}

	if strings.EqualFold(id, "") {
		return errors.Errorf("No resource idendified of type %s", ty)
	}

	dst.Spec.Resource.Name = helpers.StringPtr(name)
	dst.Spec.Resource.Id = helpers.StringPtr(id)

	dst.Status.ConditionedStatus = src.Status.ConditionedStatus
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.ManagedStatus = src.Status.ManagedStatus
	return nil
}
