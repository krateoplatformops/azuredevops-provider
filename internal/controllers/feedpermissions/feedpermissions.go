package feedpermissions

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	feedpermissionsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/feedpermissions/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/feeds"
	feedspermissions "github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/feedspermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/identities"
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
	errNotFeed             = "managed resource is not a FeedPermission custom resource"
	errUnspecifiedResource = "feed permissions resource is not specified"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(feedpermissionsv1alpha1.FeedPermissionGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(feedpermissionsv1alpha1.FeedPermissionGroupVersionKind),
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
		For(&feedpermissionsv1alpha1.FeedPermission{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*feedpermissionsv1alpha1.FeedPermission)
	if !ok {
		return nil, errors.New(errNotFeed)
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
	cr, ok := mg.(*feedpermissionsv1alpha1.FeedPermission)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotFeed)
	}

	if cr.Spec.User == nil {
		return reconciler.ExternalObservation{}, errors.New(errUnspecifiedResource)
	}

	res, err := feedspermissions.Get(ctx, e.azCli, feedspermissions.GetOptions{
		Organization: cr.Spec.Organization,
		Project:      cr.Spec.Poject,
		FeedId:       helpers.String(cr.Spec.Feed),
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	status := cr.Status.DeepCopy()

	cr.SetConditions(rtv1.Available())

	upToDate := false

	for _, feedPerm := range res.Value {
		if *feedPerm.IdentityDescriptor == status.IdentityDescriptor && feedPerm.Role == cr.Spec.User.Role {
			upToDate = true
		}
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
	cr, ok := mg.(*feedpermissionsv1alpha1.FeedPermission)
	if !ok {
		return errors.New(errNotFeed)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	spec := cr.Spec.DeepCopy()

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, &rtv1.Reference{
		Name:      *&spec.User.PojectRef.Name,
		Namespace: spec.User.PojectRef.Namespace,
	})
	if err != nil {
		return err
	}
	if project == nil {
		return errors.Errorf("Project with name %s and namespace %s not found", spec.Poject, spec.User.PojectRef.Namespace)
	}

	projectStatus := project.Status.DeepCopy()

	// match descriptor with Resource type
	userType := identities.UserType(*spec.User.Type)
	if userType != identities.BuildService {
		return errors.Errorf("identities of type %s are not supported", string(userType))
	}

	ids, err := identities.Get(ctx, e.azCli, identities.GetOptions{
		Organization: spec.Organization,
		ProjectID:    projectStatus.Id,
	})
	if err != nil {
		return err
	}

	identity, err := ids.IdentityMatch(userType, project)
	if err != nil {
		return err
	}

	cr.Status.IdentityDescriptor = identity.Descriptor

	e.kube.Status().Update(ctx, cr)

	resourceId := helpers.String(spec.Feed)
	feedPerm := []*feeds.FeedPermission{}
	feedPerm = append(feedPerm, &feeds.FeedPermission{
		DisplayName:        nil,
		IdentityDescriptor: &identity.Descriptor,
		Role:               spec.User.Role,
	})
	_, err = feedspermissions.Update(ctx, e.azCli, feedspermissions.UpdateOptions{
		Organization:    spec.Organization,
		Project:         spec.Poject,
		ResourceRole:    helpers.String(spec.User.Role),
		ResourceId:      resourceId,
		FeedPermissions: feedPerm,
	})
	if err != nil {
		e.rec.Eventf(cr, corev1.EventTypeWarning, "FeedPermissionUpdateFailed",
			"FeedPermission '%s' update failed: %s", resourceId, err.Error())
		return err
	}

	e.log.Debug("FeedPermission updated", "resource id", resourceId)

	e.rec.Eventf(cr, corev1.EventTypeNormal, "FeedPermissionUpdated",
		"FeedPermission '%s' updated", resourceId)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}
