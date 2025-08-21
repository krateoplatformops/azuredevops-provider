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
	if teamproject == nil {
		return fmt.Errorf("TeamProject with ID %s not found", src.Spec.Project)
	}
	dst.Spec.ProjectRef = &rtv1.Reference{
		Name:      teamproject.Name,
		Namespace: teamproject.Namespace,
	}
	dst.Spec.Resource = &v1alpha2.Resource{}
	if src.Spec.Resource == nil {
		return nil
	}
	dst.Spec.Resource.Type = src.Spec.Resource.Type
	if src.Spec.Resource.Id == nil {
		return nil
	}
	id := ""
	if src.Spec.Resource.Id != nil {
		id = helpers.String(src.Spec.Resource.Id)
	}
	if src.Spec.Resource.Type == nil {
		return fmt.Errorf("src resource type is nil")
	}
	ty := helpers.String(dst.Spec.Resource.Type)
	finder := resolvers.GetFinderFromType(ty)
	if finder == nil {
		return fmt.Errorf("unsupported resource type: %s", ty)
	}
	if ty == string(v1alpha2.GitRepository) {
		arr := strings.Split(id, ".")
		if len(arr) > 1 {
			id = arr[1]
		}
	}
	if id == "" {
		return fmt.Errorf("resource ID is empty for type %s", ty)
	}
	ref, err := finder(context.TODO(), cli, id)
	if err != nil {
		return err
	}
	if ref == nil {
		return fmt.Errorf("resource reference is nil for type %s and ID %s", ty, id)
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

	if src.Spec.Resource == nil {
		return fmt.Errorf("source resource is nil")
	}

	ty := helpers.String(dst.Spec.Resource.Type)
	var id, name string
	switch ty {
	case string(v1alpha2.TeamProject):
		res, err := resolvers.ResolveTeamProject(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve TeamProject: %w", err)
		}
		id = res.Status.Id
		name = res.Spec.Name
	case string(v1alpha2.GitRepository):
		res, err := resolvers.ResolveGitRepository(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve GitRepository: %w", err)
		}
		id = res.Status.Id
		name = res.Spec.Name
	case string(v1alpha2.Queue):
		res, err := resolvers.ResolveQueue(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve Queue: %w", err)
		}
		id = fmt.Sprintf("%v", helpers.Int(res.Status.Id))
		name = helpers.String(res.Spec.Name)
	case string(v1alpha2.Environment):
		res, err := resolvers.ResolveEnvironment(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve Environment: %w", err)
		}
		id = fmt.Sprintf("%v", helpers.Int(res.Status.Id))
		name = helpers.String(res.Spec.Name)
	case string(v1alpha2.Endpoint):
		res, err := resolvers.ResolveEndpoint(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve Endpoint: %w", err)
		}
		id = helpers.String(res.Status.Id)
		name = helpers.String(res.Spec.Name)
	case string(v1alpha2.VariableGroups):
		res, err := resolvers.ResolveVariableGroups(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve VariableGroups: %w", err)
		}
		id = fmt.Sprintf("%v", res.Status.Id)
		name = helpers.String(res.Spec.Name)
	case string(v1alpha2.SecureFiles):
		res, err := resolvers.ResolveSecureFiles(ctx, cli, src.Spec.Resource.ResourceRef)
		if err != nil {
			return fmt.Errorf("failed to resolve SecureFiles: %w", err)
		}
		id = helpers.String(res.Status.Id)
		name = res.Spec.Name
	default:
		return fmt.Errorf("unsupported resource type: %s", ty)
	}

	if strings.EqualFold(id, "") {
		return errors.Errorf("id is empty for your resource type %s - name: %s", ty, src.Spec.Resource.ResourceRef.Name)
	}
	if strings.EqualFold(name, "") {
		return errors.Errorf("name is empty for your resource type %s - id: %s", ty, id)
	}

	dst.Spec.Resource.Name = helpers.StringPtr(name)
	dst.Spec.Resource.Id = helpers.StringPtr(id)

	dst.Status.ConditionedStatus = src.Status.ConditionedStatus
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.ManagedStatus = src.Status.ManagedStatus
	return nil
}
