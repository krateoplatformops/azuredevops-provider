package teamproject

import (
	"context"
	"errors"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
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
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler/managed"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
)

const (
	errNotTeamProject             = "managed resource is not a TeamProject custom resource"
	annotationKeyConnectorVerbose = "krateo.io/connector-verbose"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(projects.TeamProjectGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(projects.TeamProjectGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(recorder)))

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

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return nil, errors.New(errNotTeamProject)
	}

	opts, err := c.clientOptions(ctx, cr.Spec.ConnectorConfigRef)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(cr.GetAnnotations()[annotationKeyConnectorVerbose], "true") {
		opts.Verbose = true
	}

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

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTeamProject)
	}

	if getOperationAnnotation(cr) != "" {
		op, err := e.azCli.GetOperation(ctx, azuredevops.GetOperationOpts{
			Organization: cr.Spec.Organization,
			OperationId:  getOperationAnnotation(cr),
		})
		if err != nil {
			return managed.ExternalObservation{}, resource.Ignore(httplib.IsNotFoundError, err)
		}

		if op.Status != azuredevops.StatusSucceeded {
			return managed.ExternalObservation{}, nil
		}

		prj, err := e.azCli.FindProject(ctx, azuredevops.FindProjectsOptions{
			Organization: cr.Spec.Organization,
			Name:         cr.Spec.Name,
		})
		if err != nil {
			return managed.ExternalObservation{}, err
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

		return managed.ExternalObservation{
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
			return managed.ExternalObservation{}, resource.Ignore(httplib.IsNotFoundError, err)
		}

		cr.Status.Id = helpers.String(prj.Id)
		cr.Status.Revision = *prj.Revision
		cr.Status.State = string(*prj.State)

		cr.SetConditions(rtv1.Available())

		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   false,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return errors.New(errNotTeamProject)
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
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*projects.TeamProject)
	if !ok {
		return errors.New(errNotTeamProject)
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
