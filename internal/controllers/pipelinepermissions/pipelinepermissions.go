package pipelinepermissions

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pipelinepermissionsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha1"
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
)

const (
	errNotPipeline         = "managed resource is not a PipelinePermission custom resource"
	errUnspecifiedResource = "pipeline permissions resource is not specified"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(pipelinepermissionsv1alpha1.PipelinePermissionGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(pipelinepermissionsv1alpha1.PipelinePermissionGroupVersionKind),
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
		For(&pipelinepermissionsv1alpha1.PipelinePermission{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*pipelinepermissionsv1alpha1.PipelinePermission)
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
	cr, ok := mg.(*pipelinepermissionsv1alpha1.PipelinePermission)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotPipeline)
	}

	if cr.Spec.Resource == nil {
		return reconciler.ExternalObservation{}, errors.New(errUnspecifiedResource)
	}

	res, err := pipelinespermissions.Get(ctx, e.azCli, pipelinespermissions.GetOptions{
		Organization: cr.Spec.Organization,
		Project:      cr.Spec.Project,
		ResourceType: helpers.String(cr.Spec.Resource.Type),
		ResourceId:   helpers.String(cr.Spec.Resource.Id),
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	cr.SetConditions(rtv1.Available())

	upToDate := false
	if res.AllPipelines != nil {
		current := res.AllPipelines.Authorized
		desired := helpers.BoolOrDefault(cr.Spec.Authorize, false)
		upToDate = (desired == current)
	}

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	return nil // NOOP
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*pipelinepermissionsv1alpha1.PipelinePermission)
	if !ok {
		return errors.New(errNotPipeline)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	spec := cr.Spec.DeepCopy()

	resourceId := helpers.String(spec.Resource.Id)

	_, err := pipelinespermissions.Update(ctx, e.azCli, pipelinespermissions.UpdateOptions{
		Organization: spec.Organization,
		Project:      spec.Project,
		ResourceType: helpers.String(spec.Resource.Type),
		ResourceId:   resourceId,
		ResourceAuthorization: &pipelinespermissions.ResourcePipelinePermissions{
			AllPipelines: &pipelinespermissions.Permission{
				Authorized: helpers.BoolOrDefault(spec.Authorize, false),
			},
			Pipelines: []pipelinespermissions.PipelinePermission{},
			Resource: &azuredevops.Resource{
				Id:   spec.Resource.Id,
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
		"authorize", helpers.BoolOrDefault(spec.Authorize, false))

	e.rec.Eventf(cr, corev1.EventTypeNormal, "PipelinePermissionUpdated",
		"PipelinePermission '%s' updated", resourceId)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}
