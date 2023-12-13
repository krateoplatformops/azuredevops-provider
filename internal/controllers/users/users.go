package users

import (
	"context"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	usersv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/users/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/memberships"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/users"
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
	errNotCR = "managed resource is not a Users custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(usersv1alpha1.UsersGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(usersv1alpha1.UsersGroupVersionKind),
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
		For(&usersv1alpha1.Users{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*usersv1alpha1.Users)
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
	cr, ok := mg.(*usersv1alpha1.Users)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	var user *users.UserResource
	var err error
	if cr.Status.Descriptor == nil {
		user, err = users.FindUserByName(ctx, e.azCli, users.FindUserByNameOptions{
			ListOptions: users.ListOptions{
				Organization: cr.Spec.Organization,
			},
			PrincipalName: *cr.Spec.User.Name,
		})
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		user, err = users.Get(ctx, e.azCli, users.GetOptions{
			Organization:   cr.Spec.Organization,
			UserDescriptor: helpers.String(cr.Status.Descriptor),
		})
		if err != nil && !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}
	if user == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}
	membership, err := memberships.Get(ctx, e.azCli, memberships.GetOptions{
		Organization:      cr.Spec.Organization,
		SubjectDescriptor: user.Descriptor,
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	check := true
	groupDescriptors, err := resolvers.ResolveGroupListDescriptors(ctx, e.kube, cr.Spec.GroupsRefs)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	for _, groupDescriptor := range groupDescriptors {
		err = memberships.CheckMembership(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        cr.Spec.Organization,
			SubjectDescriptor:   user.Descriptor,
			ContainerDescriptor: groupDescriptor,
		})
		if httplib.IsNotFoundError(err) {
			check = false
			break
		}
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
		check = true
	}
	if !membership.Active {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	if len(groupDescriptors) != 0 && !check {
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}

	cr.Status.Descriptor = helpers.StringPtr(user.Descriptor)

	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*usersv1alpha1.Users)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	var user *users.UserResource
	var err error
	if cr.Status.Descriptor == nil {
		user, err = users.FindUserByName(ctx, e.azCli, users.FindUserByNameOptions{
			ListOptions: users.ListOptions{
				Organization: cr.Spec.Organization,
			},
			PrincipalName: *cr.Spec.User.Name,
		})
		if err != nil {
			return err
		}
	} else {
		user, err = users.Get(ctx, e.azCli, users.GetOptions{
			Organization:   cr.Spec.Organization,
			UserDescriptor: helpers.String(cr.Status.Descriptor),
		})
		if err != nil {
			return err
		}
	}
	if user == nil {
		return errors.Errorf("user %s %s not found", helpers.String(cr.Spec.User.Name), helpers.String(cr.Spec.User.OriginID))
	}

	groupDescriptors, err := resolvers.ResolveGroupListDescriptors(ctx, e.kube, cr.Spec.GroupsRefs)
	if err != nil {
		return err
	}
	if user.OriginID == nil {
		user, err = users.Create[users.PrincipalName](ctx, e.azCli, users.CreateOptions[users.PrincipalName]{
			Organization: cr.Spec.Organization,
			Identifier: users.PrincipalName{
				PrincipalName: user.PrincipalName,
			},
			GroupDescriptors: groupDescriptors,
		})
	} else {
		user, err = users.Create[users.OriginID](ctx, e.azCli, users.CreateOptions[users.OriginID]{
			Organization: cr.Spec.Organization,
			Identifier: users.OriginID{
				OriginID: helpers.String(user.OriginID),
			},
			GroupDescriptors: groupDescriptors,
		})
	}

	if err != nil {
		return err
	}

	cr.Status.Descriptor = helpers.StringPtr(user.Descriptor)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*usersv1alpha1.Users)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	groupDescriptors, err := resolvers.ResolveGroupListDescriptors(ctx, e.kube, cr.Spec.GroupsRefs)
	if err != nil {
		return err
	}
	user, err := users.Create[users.PrincipalName](ctx, e.azCli, users.CreateOptions[users.PrincipalName]{
		Organization: cr.Spec.Organization,
		Identifier: users.PrincipalName{
			PrincipalName: helpers.String(cr.Spec.User.Name),
		},
		GroupDescriptors: groupDescriptors,
	})
	if err != nil {
		return err
	}

	cr.Status.Descriptor = helpers.StringPtr(user.Descriptor)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*usersv1alpha1.Users)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	err := users.Delete(ctx, e.azCli, users.DeleteOptions{
		Organization:   cr.Spec.Organization,
		UserDescriptor: helpers.String(cr.Status.Descriptor),
	})
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return nil
}
