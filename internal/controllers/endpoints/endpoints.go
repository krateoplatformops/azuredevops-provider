package endpoints

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	endpointsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/endpoints/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/endpoints"
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
	errNotCR = "managed resource is not a Endpoint custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(endpointsv1alpha1.EndpointGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(endpointsv1alpha1.EndpointGroupVersionKind),
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
		For(&endpointsv1alpha1.Endpoint{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
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
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	ref, err := e.resolveProjectRef(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var observed *endpoints.ServiceEndpoint
	if cr.Status.Id == nil {
		observed, err = e.findEndpoint(ctx, &ref, cr)
	} else {
		observed, err = endpoints.Get(ctx, e.azCli, endpoints.GetOptions{
			Organization: ref.Organization,
			Project:      ref.Name,
			EndpointId:   helpers.String(cr.Status.Id),
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

	cr.SetConditions(rtv1.Available())

	cr.Status.Id = helpers.StringPtr(*observed.Id)
	cr.Status.Url = helpers.StringPtr(*observed.Url)

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Status().Update(ctx, cr)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	ref, err := e.resolveProjectRef(ctx, cr)
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Creating())

	res, err := endpoints.Create(ctx, e.azCli, endpoints.CreateOptions{
		Organization: ref.Organization,
		Endpoint:     asAzureDevopsServiceEndpoint(&ref, cr),
	})
	if err != nil {
		return err
	}

	cr.Status.Id = helpers.StringPtr(*res.Id)
	cr.Status.Url = helpers.StringPtr(*res.Url)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // NOOP
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	if cr.Status.Id == nil {
		return fmt.Errorf("missing Queue identifier")
	}

	ref, err := e.resolveProjectRef(ctx, cr.DeepCopy())
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return endpoints.Delete(ctx, e.azCli, endpoints.DeleteOptions{
		Organization: ref.Organization,
		ProjectIds:   []string{ref.Name},
		EndpointId:   helpers.String(cr.Status.Id),
	})
}
