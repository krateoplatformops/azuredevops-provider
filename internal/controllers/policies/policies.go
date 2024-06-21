package policies

import (
	"context"
	"errors"
	"fmt"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/lucasepe/httplib"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/event"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/policies"

	policiesv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/policies/v1alpha1"
)

const (
	errNotPolicy = "managed resource is not a Policy custom resource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(policiesv1alpha1.PolicyGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(policiesv1alpha1.PolicyGroupVersionKind),
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
		For(&policiesv1alpha1.Policy{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*policiesv1alpha1.Policy)
	if !ok {
		return nil, errors.New(errNotPolicy)
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
	cr, ok := mg.(*policiesv1alpha1.Policy)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotPolicy)
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PolicyBody.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, fmt.Errorf("failed to resolve project reference: %v", err)
	}

	response := &policies.PolicyBody{}

	if cr.Status.ID == nil && cr.Spec.PolicyBody.ID != nil {
		response, err = policies.Find(ctx, e.azCli, policies.FindOptions{
			Organization:    project.Spec.Organization,
			ProjectId:       project.Status.Id,
			ConfigurationId: *cr.Spec.PolicyBody.ID,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	} else if cr.Status.ID != nil {
		response, err = policies.Get(ctx, e.azCli, policies.GetOptions{
			Organization:    project.Spec.Organization,
			ProjectId:       project.Status.Id,
			ConfigurationId: *cr.Status.ID,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}
	if response == nil || response.IsDeleted {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	cr.Status.ID = &response.ID
	cr.Status.URL = &response.URL
	cr.SetConditions(rtv1.Available())

	err = e.kube.Status().Update(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	if !isUpdated(ctx, e.kube, cr, response) {
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*policiesv1alpha1.Policy)
	if !ok {
		return errors.New(errNotPolicy)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PolicyBody.ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve project reference: %v", err)
	}

	policy, err := customResourceToPolicy(ctx, e.kube, cr)
	if err != nil {
		return fmt.Errorf("failed to convert custom resource to : %v", err)
	}

	response, err := policies.Create(ctx, e.azCli, policies.CreateOptions{
		PolicyBody:   policy,
		Organization: project.Spec.Organization,
		ProjectId:    project.Status.Id,
	})

	if err != nil {
		return fmt.Errorf("failed to create Policy: %v", err)
	}

	cr.SetConditions(rtv1.Creating())
	cr.Status.ID = &response.ID
	cr.Status.URL = &response.URL

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*policiesv1alpha1.Policy)
	if !ok {
		return errors.New(errNotPolicy)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PolicyBody.ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve project reference: %v", err)
	}

	policy, err := customResourceToPolicy(ctx, e.kube, cr)
	if err != nil {
		return fmt.Errorf("failed to convert custom resource to : %v", err)
	}

	response, err := policies.Update(ctx, e.azCli, policies.UpdateOptions{
		PolicyBody:      policy,
		Organization:    project.Spec.Organization,
		ProjectId:       project.Status.Id,
		ConfigurationId: *cr.Status.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to update Policy: %v", err)
	}

	cr.SetConditions(rtv1.Creating())
	cr.Status.ID = &response.ID
	cr.Status.URL = &response.URL

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*policiesv1alpha1.Policy)
	if !ok {
		return errors.New(errNotPolicy)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PolicyBody.ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve project reference: %v", err)
	}

	err = policies.Delete(ctx, e.azCli, policies.DeleteOptions{
		Organization:    project.Spec.Organization,
		ProjectId:       project.Status.Id,
		ConfigurationId: *cr.Status.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete Policy: %v", err)
	}

	return nil
}
