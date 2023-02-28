package teamproject

import (
	"context"
	"errors"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/lucasepe/httplib"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/event"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
)

const (
	errNotTeamProject = "managed resource is not a TeamProject custom resource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(projects.TeamProjectGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(projects.TeamProjectGroupVersionKind),
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
		For(&projects.TeamProject{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return nil, errors.New(errNotTeamProject)
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
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotTeamProject)
	}

	if getOperationAnnotation(cr) != "" {
		op, err := e.azCli.GetOperation(ctx, azuredevops.GetOperationOpts{
			Organization: cr.Spec.Organization,
			OperationId:  getOperationAnnotation(cr),
		})
		if err != nil {
			return reconciler.ExternalObservation{}, resource.Ignore(httplib.IsNotFoundError, err)
		}

		if op.Status != azuredevops.StatusSucceeded {
			return reconciler.ExternalObservation{}, nil
		}

		prj, err := e.azCli.FindProject(ctx, azuredevops.FindProjectsOptions{
			Organization: cr.Spec.Organization,
			Name:         cr.Spec.Name,
		})
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}

		e.log.Debug("Found Project", "id", *prj.Id, "name", prj.Name)

		deleteOperationAnnotation(cr)
		meta.SetExternalName(cr, helpers.String(prj.Id))

		cr.Status.Id = helpers.String(prj.Id)
		cr.Status.Revision = *prj.Revision
		cr.Status.State = string(*prj.State)

		cr.SetConditions(rtv1.Available())

		//e.rec.Eventf(cr, corev1.EventTypeNormal, "TeamProjectCreated",
		//	"TeamProject '%s/%s' created", cr.Spec.Org, cr.Spec.Name)

		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, e.kube.Update(ctx, cr)
	}

	if meta.GetExternalName(cr) != "" {
		prj, err := e.azCli.GetProject(ctx, azuredevops.GetProjectOptions{
			Organization: cr.Spec.Organization,
			ProjectId:    meta.GetExternalName(cr),
		})
		if err != nil {
			return reconciler.ExternalObservation{}, resource.Ignore(httplib.IsNotFoundError, err)
		}

		cr.Status.Id = helpers.String(prj.Id)
		cr.Status.Revision = *prj.Revision
		cr.Status.State = string(*prj.State)

		cr.SetConditions(rtv1.Available())

		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	return reconciler.ExternalObservation{
		ResourceExists:   false,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return errors.New(errNotTeamProject)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	if getOperationAnnotation(cr) != "" {
		return nil
	}

	cr.SetConditions(rtv1.Creating())

	spec := cr.Spec.DeepCopy()

	op, err := e.azCli.CreateProject(ctx, azuredevops.CreateProjectOptions{
		Organization: spec.Organization,
		TeamProject:  teamProjectFromSpec(spec),
	})
	if err != nil {
		return err
	}

	setOperationAnnotation(cr, op.Id)
	cr.SetConditions(conditionFromOperationReference(op))

	e.log.Debug("Creating TeamProject", "org", spec.Organization, "name", spec.Name, "status", op.Status)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "TeamProjectCreating",
		"TeamProject '%s/%s' creating", spec.Organization, spec.Name)

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	//if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
	//	e.log.Debug("External resource should not be updated by provider, skip updating.")
	//	return nil
	//}

	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return errors.New(errNotTeamProject)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	cr.SetConditions(rtv1.Deleting())

	_, err := e.azCli.DeleteProject(ctx, azuredevops.DeleteProjectOptions{
		Organization: cr.Spec.Organization,
		ProjectId:    cr.Status.Id,
	})
	if err != nil {
		return resource.Ignore(httplib.IsNotFoundError, err)
	}

	e.log.Debug("TeamProject deleted",
		"id", cr.Status.Id, "org", cr.Spec.Organization, "name", cr.Spec.Name)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "TeamProjectDeleted",
		"TeamProject '%s/%s' deleted", cr.Spec.Organization, cr.Spec.Name)

	return nil
}
