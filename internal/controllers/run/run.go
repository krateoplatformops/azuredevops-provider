package run

import (
	"context"
	"fmt"
	"strconv"

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

	runsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/runs/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/runs"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
)

const (
	errNotCR = "managed resource is not a Run custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(runsv1alpha1.RunGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(runsv1alpha1.RunGroupVersionKind),
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
		For(&runsv1alpha1.Run{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*runsv1alpha1.Run)
	if !ok {
		return nil, errors.New(errNotCR)
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
	cr, ok := mg.(*runsv1alpha1.Run)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	spec := cr.Spec.DeepCopy()

	pip, err := resolvers.ResolvePipeline(ctx, e.kube, spec.PipelineRef)
	if err != nil || pip == nil {
		return reconciler.ExternalObservation{},
			errors.Wrapf(err, "unble to resolve Pipeline: %s", spec.PipelineRef.Name)
	}

	pipelineId, err := strconv.Atoi(helpers.String(pip.Status.Id))
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, pip.Spec.ProjectRef)
	if err != nil || prj == nil {
		return reconciler.ExternalObservation{},
			errors.Wrapf(err, "unble to resolve Project: %s", pip.Spec.ProjectRef.Name)
	}

	var run *runs.RunInfo
	if runId := meta.GetExternalName(cr); runId != "" {
		id, err := strconv.Atoi(runId)
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}

		run, err = runs.Get(ctx, e.azCli, runs.GetOptions{
			Organization: prj.Spec.Organization,
			Project:      prj.Status.Id,
			PipelineId:   pipelineId,
			RunId:        id,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if run == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	cr.SetConditions(rtv1.Available())

	cr.Status.Id = helpers.IntPtr(*run.Id)
	cr.Status.PipelineId = helpers.IntPtr(pipelineId)
	cr.Status.State = helpers.StringPtr(*run.State)
	cr.Status.Url = helpers.StringPtr(*run.Url)

	//if err := e.kube.Update(ctx, cr); err != nil {
	//	return reconciler.ExternalObservation{}, err
	//}

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*runsv1alpha1.Run)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	cr.SetConditions(rtv1.Creating())

	spec := cr.Spec.DeepCopy()

	pip, err := resolvers.ResolvePipeline(ctx, e.kube, spec.PipelineRef)
	if err != nil || pip == nil {
		return errors.Wrapf(err, "unble to resolve Pipeline: %s", spec.PipelineRef.Name)
	}

	pipelineId, err := strconv.Atoi(helpers.String(pip.Status.Id))
	if err != nil {
		return err
	}

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, pip.Spec.ProjectRef)
	if err != nil || prj == nil {
		return errors.Wrapf(err, "unble to resolve Project: %s", pip.Spec.ProjectRef.Name)
	}

	if cr.Spec.RunParameters.Resources == nil {
		fmt.Println("RunParameters is nil")
	}

	run, err := runs.Run(ctx, e.azCli, runs.RunOptions{
		Organization: prj.Spec.Organization,
		Project:      prj.Status.Id,
		PipelineId:   pipelineId,
		RunParameters: &runs.RunPipelineParameters{
			PreviewRun:   cr.Spec.RunParameters.PreviewRun,
			StagesToSkip: cr.Spec.RunParameters.StagesToSkip,
			Resources: &runs.RunResourcesParameters{
				Repositories: cr.Spec.RunParameters.Resources.Repositories,
				Builds:       cr.Spec.RunParameters.Resources.Builds,
				Containers:   cr.Spec.RunParameters.Resources.Containers,
				Packages:     cr.Spec.RunParameters.Resources.Packages,
				Pipelines:    cr.Spec.RunParameters.Resources.Pipelines,
			},
			TemplateParameters: cr.Spec.RunParameters.TemplateParameters,
			Variables:          cr.Spec.RunParameters.Variables,
			YamlOverride:       cr.Spec.RunParameters.YamlOverride,
		}})
	if err != nil {
		return err
	}

	runId := fmt.Sprintf("%d", helpers.Int(run.Id))
	meta.SetExternalName(cr, runId)

	cr.Status.Id = helpers.IntPtr(*run.Id)
	cr.Status.PipelineId = helpers.IntPtr(pipelineId)
	cr.Status.State = helpers.StringPtr(*run.State)
	cr.Status.Url = helpers.StringPtr(*run.Url)

	if err := e.kube.Update(ctx, cr); err != nil {
		return err
	}

	e.log.Debug("Pipeline run issued", "id", runId, "url", helpers.String(run.Url))
	e.rec.Eventf(cr, corev1.EventTypeNormal, "PipelineRunIssued",
		"Run '%s' issued", helpers.String(run.Url))

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}
