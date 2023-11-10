package environments

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"

	environmentsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/environments/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/environments"
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
	"github.com/lucasepe/httplib"
	"github.com/pkg/errors"
)

const (
	errNotCR = "managed resource is not a Environment custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(environmentsv1alpha1.EnvironmentGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(environmentsv1alpha1.EnvironmentGroupVersionKind),
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
		For(&environmentsv1alpha1.Environment{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*environmentsv1alpha1.Environment)
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
	cr, ok := mg.(*environmentsv1alpha1.Environment)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var observed *environments.Environment
	if cr.Status.Id == nil {
		observed, err = e.findEnvironment(ctx, cr)
	} else {
		observed, err = environments.Get(ctx, e.azCli, environments.GetOptions{
			Organization:  organization,
			Project:       project,
			Description:   helpers.String(cr.Spec.Description),
			Name:          helpers.String(cr.Spec.Name),
			EnvironmentId: helpers.Int(cr.Status.Id),
		})
	}
	if err != nil {
		if !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if observed == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}
	if helpers.String(observed.Name) != helpers.String(cr.Spec.Name) || helpers.String(observed.Description) != helpers.String(cr.Spec.Description) {
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}

	cr.SetConditions(rtv1.Available())

	cr.Status.Id = observed.Id

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Status().Update(ctx, cr)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*environmentsv1alpha1.Environment)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Creating())

	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	res, err := environments.Create(ctx, e.azCli, environments.CreateOptions{
		Organization: organization,
		Project:      project,
		Environment: &environments.Environment{
			Name:        helpers.StringPtr(name),
			Description: cr.Spec.Description,
		},
	})
	if err != nil {
		return err
	}

	cr.Status.Id = res.Id

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*environmentsv1alpha1.Environment)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	environmentId := helpers.IntOrDefault(cr.Status.Id, -1)
	if helpers.Int(environmentId) == -1 {
		return fmt.Errorf("missing Environment identifier")
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr.DeepCopy())
	if err != nil {
		return err
	}

	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	_, err = environments.Update(ctx, e.azCli, environments.UpdateOptions{
		Organization:  organization,
		Project:       project,
		EnvironmentId: helpers.Int(environmentId),
		Environment: &environments.Environment{
			Name:        helpers.StringPtr(name),
			Description: cr.Spec.Description,
		},
	})

	return err
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*environmentsv1alpha1.Environment)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}
	cr.SetConditions(rtv1.Deleting())

	environmentId := helpers.IntOrDefault(cr.Status.Id, -1)
	if helpers.Int(environmentId) == -1 {
		return fmt.Errorf("missing Environment identifier")
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr.DeepCopy())
	if err != nil {
		return err
	}

	err = environments.Delete(ctx, e.azCli, environments.DeleteOptions{
		Organization:  organization,
		Project:       project,
		EnvironmentId: helpers.Int(environmentId),
	})
	if err != nil {
		return resource.Ignore(httplib.IsNotFoundError, err)
	}

	e.log.Debug("Environment deleted",
		"id", cr.Status.Id, "org", cr.Spec.ProjectRef.Namespace, "project", cr.Spec.ProjectRef.Name, "name", cr.Spec.Name)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "EnvironmentDeleted",
		"Environment '%s' in '%s/%s' deleted", helpers.String(cr.Spec.Name), cr.Spec.ProjectRef.Namespace, cr.Spec.ProjectRef.Name)
	return nil
}

func (e *external) resolveProjectAndOrg(ctx context.Context, cr *environmentsv1alpha1.Environment) (string, string, error) {
	var project, organization string
	if cr.Spec.ProjectRef != nil {
		prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
		if err != nil {
			return "", "", errors.Wrapf(err, "unable to resolve TeamProject: %s", cr.Spec.ProjectRef.Name)
		}
		if prj != nil {
			project = prj.Spec.Name
			organization = prj.Spec.Organization
		}
	}

	if len(project) == 0 {
		return "", "", fmt.Errorf("missing Project name")
	}

	if len(organization) == 0 {
		return "", "", fmt.Errorf("missing Organization name")
	}

	return organization, project, nil
}

func (e *external) findEnvironment(ctx context.Context, cr *environmentsv1alpha1.Environment) (*environments.Environment, error) {
	org, prj, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return nil, err
	}

	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	return environments.Find(ctx, e.azCli, environments.FindOptions{
		Organization:    org,
		Project:         prj,
		EnvironmentName: name,
	})
}
