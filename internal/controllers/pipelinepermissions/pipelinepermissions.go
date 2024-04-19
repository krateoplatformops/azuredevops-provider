package pipelinepermissions

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinepermissionsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha1"
	pipelinepermissionsv1alpha2 "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha2"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	pipelinespermissions "github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/pipelinespermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/event"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	errNotPipeline         = "managed resource is not a PipelinePermission custom resource"
	errUnspecifiedResource = "pipeline permissions resource is not specified"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(pipelinepermissionsv1alpha2.PipelinePermissionGroupKind)

	log := o.Logger.WithValues("controller", name)

	if err := (&pipelinepermissionsv1alpha2.PipelinePermission{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create webhook webhook PipelinePermission: %s", err)
	}
	if err := (&pipelinepermissionsv1alpha1.PipelinePermission{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create webhook webhook PipelinePermission: %s", err)
	}
	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(pipelinepermissionsv1alpha2.PipelinePermissionGroupVersionKind),
		reconciler.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		reconciler.WithPollInterval(o.PollInterval),
		reconciler.WithLogger(log),
		reconciler.WithRecorder(event.NewAPIRecorder(recorder)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&pipelinepermissionsv1alpha2.PipelinePermission{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*pipelinepermissionsv1alpha2.PipelinePermission)
	if !ok {
		return nil, errors.New(errNotPipeline)
	}

	opts, err := resolvers.ResolveConnectorConfig(ctx, c.kube, cr.Spec.ConnectorConfigRef)
	if err != nil {
		return nil, err
	}

	opts.Verbose = meta.IsVerbose(cr)

	return &external{
		kube:  c.kube,
		log:   c.log,
		azCli: azuredevops.NewClient(opts),
		rec:   c.recorder,
	}, nil
}

type external struct {
	kube  client.Client
	log   logging.Logger
	azCli *azuredevops.Client
	rec   record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (reconciler.ExternalObservation, error) {
	cr, ok := mg.(*pipelinepermissionsv1alpha2.PipelinePermission)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotPipeline)
	}

	if cr.Spec.Resource == nil {
		return reconciler.ExternalObservation{}, errors.New(errUnspecifiedResource)
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	resourceId, err := resolveResourceId(ctx, e.kube, cr.Spec.Resource.ResourceRef, helpers.String(cr.Spec.Resource.Type))
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	res, err := pipelinespermissions.Get(ctx, e.azCli, pipelinespermissions.GetOptions{
		Organization: project.Spec.Organization,
		Project:      project.Status.Id,
		ResourceType: helpers.String(cr.Spec.Resource.Type),
		ResourceId:   helpers.String(resourceId),
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	cr.SetConditions(rtv1.Available())

	if res.AllPipelines != nil {
		current := res.AllPipelines.Authorized
		desired := helpers.BoolOrDefault(cr.Spec.AuthorizeAll, false)
		if !(desired == current) {
			return reconciler.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: false,
			}, nil
		}
	}

	ok, err = checkPipelinePermission(ctx, e.kube, cr.Spec.Pipelines, res.Pipelines)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: ok,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	return nil // NOOP
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*pipelinepermissionsv1alpha2.PipelinePermission)
	if !ok {
		return errors.New(errNotPipeline)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	spec := cr.Spec.DeepCopy()

	teamproject, err := resolvers.ResolveTeamProject(ctx, e.kube, spec.ProjectRef)
	if err != nil {
		return err
	}

	resourceId, err := resolveResourceId(ctx, e.kube, cr.Spec.Resource.ResourceRef, helpers.String(cr.Spec.Resource.Type))
	if err != nil {
		return err
	}

	pipelineList := []pipelinespermissions.PipelinePermission{}

	for _, v := range cr.Spec.Pipelines {
		pipeline, err := resolvers.ResolvePipeline(ctx, e.kube, v.PipelineRef)
		if err != nil {
			return err
		}
		pipelineList = append(pipelineList, pipelinespermissions.PipelinePermission{
			Authorized: v.Authorized,
			Id:         pipeline.Status.Id,
		})
	}

	_, err = pipelinespermissions.Update(ctx, e.azCli, pipelinespermissions.UpdateOptions{
		Organization: teamproject.Spec.Organization,
		Project:      teamproject.Status.Id,
		ResourceType: helpers.String(spec.Resource.Type),
		ResourceId:   helpers.String(resourceId),
		ResourceAuthorization: &pipelinespermissions.ResourcePipelinePermissions{
			AllPipelines: &pipelinespermissions.Permission{
				Authorized: helpers.BoolOrDefault(spec.AuthorizeAll, false),
			},
			Pipelines: pipelineList,
			Resource: &azuredevops.Resource{
				Id:   resourceId,
				Type: spec.Resource.Type,
			},
		},
	})
	if err != nil {
		e.rec.Eventf(cr, corev1.EventTypeWarning, "PipelinePermissionUpdateFailed",
			"PipelinePermission '%s' update failed: %s", resourceId, err.Error())
		return err
	}

	e.log.Debug("PipelinePermission updated", "resource id", resourceId,
		"authorize", helpers.BoolOrDefault(spec.AuthorizeAll, false))

	e.rec.Eventf(cr, corev1.EventTypeNormal, "PipelinePermissionUpdated",
		"PipelinePermission '%s' updated", resourceId)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func resolveResourceId(ctx context.Context, cli client.Client, ref *rtv1.Reference, ty string) (*string, error) {
	if ref == nil {
		return nil, fmt.Errorf("no resource referenced")
	}
	ty = strings.ToLower(ty)
	switch ty {
	case string(pipelinepermissionsv1alpha2.GitRepository):
		repo, err := resolvers.ResolveGitRepository(ctx, cli, ref)
		if err != nil {
			return nil, err
		}
		proj, err := resolvers.ResolveTeamProject(ctx, cli, repo.Spec.ProjectRef)
		ret := fmt.Sprintf("%s.%s", proj.Status.Id, repo.Status.Id)
		return helpers.StringPtr(ret), err
	case string(pipelinepermissionsv1alpha2.Environment):
		env, err := resolvers.ResolveEnvironment(ctx, cli, ref)
		ret := fmt.Sprintf("%v", helpers.Int(env.Status.Id))
		return helpers.StringPtr(ret), err
	case string(pipelinepermissionsv1alpha2.Queue):
		que, err := resolvers.ResolveQueue(ctx, cli, ref)
		ret := fmt.Sprintf("%v", helpers.Int(que.Status.Id))
		return helpers.StringPtr(ret), err
	case string(pipelinepermissionsv1alpha2.Endpoint):
		end, err := resolvers.ResolveEndpoint(ctx, cli, ref)
		ret := helpers.String(end.Status.Id)
		return helpers.StringPtr(ret), err
	}

	return nil, fmt.Errorf("no resource referenced of type %s", ty)
}
