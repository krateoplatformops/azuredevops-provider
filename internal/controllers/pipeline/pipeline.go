package pipeline

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/event"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"
	"github.com/lucasepe/httplib"
	"github.com/pkg/errors"

	pipelinesv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/pipelines/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	pipelines "github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/pipelines"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
)

const (
	errNotPipeline = "managed resource is not a Pipeline custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(pipelinesv1alpha1.PipelineGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(pipelinesv1alpha1.PipelineGroupVersionKind),
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
		For(&pipelinesv1alpha1.Pipeline{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*pipelinesv1alpha1.Pipeline)
	if !ok {
		return nil, errors.New(errNotPipeline)
	}

	opts, err := resolvers.ResolveConnectorConfig(ctx, c.kube, cr.Spec.ConnectorConfigRef)
	if err != nil {
		return nil, err
	}

	opts.Verbose = meta.IsVerbose(cr)

	log := c.log.WithValues("name", cr.Name, "apiVersion", cr.APIVersion, "kind", cr.Kind)

	return &external{
		kube:  c.kube,
		log:   log,
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
	cr, ok := mg.(*pipelinesv1alpha1.Pipeline)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotPipeline)
	}

	spec := cr.Spec.DeepCopy()

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, spec.ProjectRef)
	if err != nil || prj == nil {
		return reconciler.ExternalObservation{},
			errors.Wrapf(err, "unble to resolve TeamProject: %s", spec.ProjectRef.Name)
	}

	var pip *pipelines.Pipeline

	if pipId := meta.GetExternalName(cr); pipId != "" {
		var err error
		pip, err = pipelines.Get(ctx, e.azCli, pipelines.GetOptions{
			Organization: prj.Spec.Organization,
			Project:      prj.Status.Id,
			PipelineId:   pipId,
		})
		if err != nil && !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if pip == nil {
		var err error
		pip, err = pipelines.Find(ctx, e.azCli, pipelines.FindOptions{
			Organization: prj.Spec.Organization,
			Project:      prj.Spec.Name,
			Name:         spec.Name,
		})
		if err != nil && !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if pip == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	if pip.Id == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	pipId := fmt.Sprintf("%d", *pip.Id)
	meta.SetExternalName(cr, pipId)
	if err := e.kube.Update(ctx, cr); err != nil {
		return reconciler.ExternalObservation{}, err
	}

	cr.Status.Id = helpers.StringPtr(pipId)
	cr.Status.Url = helpers.StringPtr(*pip.Url)

	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*pipelinesv1alpha1.Pipeline)
	if !ok {
		return errors.New(errNotPipeline)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	cr.SetConditions(rtv1.Creating())

	e.log.Info("Creating resource")

	spec := cr.Spec.DeepCopy()

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, spec.ProjectRef)
	if err != nil {
		return errors.Wrapf(err, "unble to resolve TeamProject: %s", spec.ProjectRef.Name)
	}

	repo, err := resolvers.ResolveGitRepository(ctx, e.kube, spec.RepositoryRef)
	if err != nil {
		return errors.Wrapf(err, "unable to resolve GitRepository: %s", spec.RepositoryRef.Name)
	}

	res, err := pipelines.Create(ctx, e.azCli, pipelines.CreateOptions{
		Organization: prj.Spec.Organization,
		Project:      prj.Status.Id,
		Pipeline: pipelines.Pipeline{
			Folder: spec.Folder,
			Name:   spec.Name,
			Configuration: &pipelines.PipelineConfiguration{
				Type: pipelines.ConfigurationType(*spec.ConfigurationType),
				Path: spec.DefinitionPath,
				Repository: &pipelines.BuildRepository{
					Id:   repo.Status.Id,
					Name: repo.Spec.Name,
					Type: pipelines.BuildRepositoryType(*spec.RepositoryType),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	pipelineId := fmt.Sprintf("%d", *res.Id)
	meta.SetExternalName(cr, pipelineId)
	if err := e.kube.Update(ctx, cr); err != nil {
		return err
	}

	e.log.Debug("Pipeline created", "id", pipelineId, "url", helpers.String(res.Url))
	e.rec.Eventf(cr, corev1.EventTypeNormal, "PipelineCreated",
		"Pipeline '%s' created", helpers.String(res.Url))

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*pipelinesv1alpha1.Pipeline)
	if !ok {
		return errors.New(errNotPipeline)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}
	cr.SetConditions(rtv1.Deleting())

	e.log.Info("Deleting resource")

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return errors.Wrapf(err, "unble to resolve TeamProject: %s", cr.Spec.ProjectRef.Name)
	}

	err = e.azCli.DeleteDefinition(ctx, azuredevops.DeleteDefinitionOptions{
		Organization: prj.Spec.Organization,
		Project:      prj.Status.Id,
		DefinitionId: helpers.String(cr.Status.Id),
	})
	if err != nil {
		return resource.Ignore(httplib.IsNotFoundError, err)
	}

	e.log.Debug("Pipeline deleted", "id", cr.Status.Id, "url", helpers.String(cr.Status.Url))
	e.rec.Eventf(cr, corev1.EventTypeNormal, "PipelineDeleted",
		"Pipeline '%s' deleted", helpers.String(cr.Status.Url))

	return nil // noop
}
