package users

import (
	"context"
	"fmt"

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
			PrincipalName: helpers.String(cr.Spec.User.Name),
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
	// group and team descriptors are managed by as the same object in azure devops APIs
	groupAndTeamDescriptors, err := resolvers.ResolveGroupAndTeamDescriptors(ctx, e.kube, cr.Spec.GroupsRefs, cr.Spec.TeamsRefs)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	for _, descriptor := range groupAndTeamDescriptors {
		err = memberships.CheckMembership(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        cr.Spec.Organization,
			SubjectDescriptor:   user.Descriptor,
			ContainerDescriptor: descriptor,
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

	if len(groupAndTeamDescriptors) != 0 && !check {
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

	e.log.Info("Updating resource")

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

	// group and team descriptors are managed by as the same object in azure devops APIs
	groupAndTeamDescriptors, err := resolvers.ResolveGroupAndTeamDescriptors(ctx, e.kube, cr.Spec.GroupsRefs, cr.Spec.TeamsRefs)
	if err != nil {
		return err
	}

	for _, descriptor := range groupAndTeamDescriptors {
		err = memberships.Create(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        cr.Spec.Organization,
			SubjectDescriptor:   user.Descriptor,
			ContainerDescriptor: descriptor,
		})
		if err != nil {
			return fmt.Errorf("failed to add user %s to group or team: %w", user.PrincipalName, err)
		}
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

	e.log.Info("Creating resource")

	// group and team descriptors are managed by as the same object in azure devops APIs
	groupAndTeamDescriptors, err := resolvers.ResolveGroupAndTeamDescriptors(ctx, e.kube, cr.Spec.GroupsRefs, cr.Spec.TeamsRefs)
	if err != nil {
		return err
	}

	var user *users.UserResource
	if cr.Spec.User.OriginID != nil {
		// User is assumed to be an Azure Active Directory user
		user, err = users.Create[users.OriginID](ctx, e.azCli, users.CreateOptions[users.OriginID]{
			Organization: cr.Spec.Organization,
			Identifier: users.OriginID{
				OriginID: helpers.String(cr.Spec.User.OriginID),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", helpers.String(cr.Spec.User.OriginID), err)
		}
	} else {
		// User could be an Azure DevOps user or an Azure Active Directory user - this api will create an Azure DevOps user if it does not exist the corresponding Azure Active Directory user
		user, err = users.Create[users.PrincipalName](ctx, e.azCli, users.CreateOptions[users.PrincipalName]{
			Organization: cr.Spec.Organization,
			Identifier: users.PrincipalName{
				PrincipalName: helpers.String(cr.Spec.User.Name),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", helpers.String(cr.Spec.User.Name), err)
		}
	}

	for _, descriptor := range groupAndTeamDescriptors {
		err = memberships.Create(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        cr.Spec.Organization,
			SubjectDescriptor:   user.Descriptor,
			ContainerDescriptor: descriptor,
		})
		if err != nil {
			return fmt.Errorf("failed to add user %s to group or team: %w", user.PrincipalName, err)
		}
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

	e.log.Info("Deleting resource")

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
