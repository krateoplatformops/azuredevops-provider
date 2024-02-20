package groups

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	groupsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/groups/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/descriptors"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/groups"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/memberships"
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
	errNotCR = "managed resource is not a Groups custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(groupsv1alpha1.GroupsGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(groupsv1alpha1.GroupsGroupVersionKind),
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
		For(&groupsv1alpha1.Groups{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*groupsv1alpha1.Groups)
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

func resolveProjectAndOrganization(ctx context.Context, kube client.Client, cr *groupsv1alpha1.Groups) (*string, *string, error) {
	var projectId, organization *string
	if cr.Spec.Membership.ProjectRef == nil {
		if cr.Spec.Membership.Organization == nil {
			return nil, nil, errors.New("spec.membership.organization or spec.membership.projectRef must be set")
		}
		organization = cr.Spec.Membership.Organization
		return nil, organization, nil
	}
	project, err := resolvers.ResolveTeamProject(ctx, kube, cr.Spec.Membership.ProjectRef)
	if err != nil {
		return nil, nil, err
	}
	projectId = &project.Status.Id
	organization = &project.Spec.Organization
	return projectId, organization, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (reconciler.ExternalObservation, error) {
	cr, ok := mg.(*groupsv1alpha1.Groups)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}
	projectId, organization, err := resolveProjectAndOrganization(ctx, e.kube, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var group *groups.GroupResponse
	if cr.Status.Descriptor == nil {
		group, err = groups.FindGroupByName(ctx, e.azCli, groups.FindGroupByNameOptions{
			ListOptions: groups.ListOptions{
				Organization: helpers.String(organization),
			},
			GroupName: helpers.String(cr.Spec.GroupsName),
			ProjectID: helpers.String(projectId),
		})
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		group, err = groups.Get(ctx, e.azCli, groups.GetOptions{
			Organization:    helpers.String(organization),
			GroupDescriptor: *cr.Status.Descriptor,
		})
		if err != nil && !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if group == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}
	membership, err := memberships.Get(ctx, e.azCli, memberships.GetOptions{
		Organization:      helpers.String(organization),
		SubjectDescriptor: group.Descriptor,
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	if !membership.Active {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}
	check := true

	// group and team descriptors are managed by as the same object in azure devops APIs
	groupAndTeamDescriptors, err := resolvers.ResolveGroupAndTeamDescriptors(ctx, e.kube, cr.Spec.GroupsRefs, cr.Spec.TeamsRefs)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	for _, descriptor := range groupAndTeamDescriptors {
		err = memberships.CheckMembership(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   group.Descriptor,
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

	if len(groupAndTeamDescriptors) != 0 && !check {
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}

	cr.Status.Descriptor = helpers.StringPtr(group.Descriptor)

	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*groupsv1alpha1.Groups)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	projectId, organization, err := resolveProjectAndOrganization(ctx, e.kube, cr)
	if err != nil {
		return err
	}

	var group *groups.GroupResponse
	if cr.Status.Descriptor == nil {
		group, err = groups.FindGroupByName(ctx, e.azCli, groups.FindGroupByNameOptions{
			ListOptions: groups.ListOptions{
				Organization: helpers.String(organization),
			},
			ProjectID: helpers.String(projectId),
			GroupName: helpers.String(cr.Spec.GroupsName),
		})
		if err != nil {
			return err
		}
	} else {
		group, err = groups.Get(ctx, e.azCli, groups.GetOptions{
			Organization:    helpers.String(organization),
			GroupDescriptor: helpers.String(cr.Status.Descriptor),
		})
		if err != nil {
			return err
		}
	}
	if group == nil {
		return errors.Errorf("group %s %s not found", *cr.Spec.GroupsName, *cr.Spec.OriginID)
	}

	var projectDescriptor *descriptors.DescriptorResponse
	if projectId != nil {
		projectDescriptor, err = descriptors.Get(ctx, e.azCli, descriptors.GetOptions{
			Organization: helpers.String(organization),
			ResourceID:   helpers.String(projectId),
		})
		if err != nil {
			return err
		}
	}
	var scopeDescriptor *string
	if projectDescriptor != nil {
		scopeDescriptor = projectDescriptor.Value
	}

	// group and team descriptors are managed by as the same object in azure devops APIs
	groupAndTeamDescriptors, err := resolvers.ResolveGroupAndTeamDescriptors(ctx, e.kube, cr.Spec.GroupsRefs, cr.Spec.TeamsRefs)
	if err != nil {
		return err
	}

	group, err = groups.Create(ctx, e.azCli, groups.CreateOptions[groups.GroupDescription]{
		Organization:    helpers.String(organization),
		ScopeDescriptor: scopeDescriptor,
		GroupData: groups.GroupDescription{
			DisplayName: helpers.String(cr.Spec.GroupsName),
			Description: cr.Spec.Description,
		}})
	if err != nil {
		return err
	}

	for _, containerDescriptor := range groupAndTeamDescriptors {
		err = memberships.Create(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   group.Descriptor,
			ContainerDescriptor: containerDescriptor,
		})
		if err != nil {
			return fmt.Errorf("failed to create group membership: %w", err)
		}
	}

	cr.Status.Descriptor = helpers.StringPtr(group.Descriptor)

	return e.kube.Status().Update(ctx, cr)
}
func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*groupsv1alpha1.Groups)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	projectId, organization, err := resolveProjectAndOrganization(ctx, e.kube, cr)
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Creating())

	var projectDescriptor *descriptors.DescriptorResponse
	if projectId != nil {
		projectDescriptor, err = descriptors.Get(ctx, e.azCli, descriptors.GetOptions{
			Organization: helpers.String(organization),
			ResourceID:   helpers.String(projectId),
		})
		if err != nil {
			return err
		}
	}
	var scopeDescriptor *string
	if projectDescriptor != nil {
		scopeDescriptor = projectDescriptor.Value
	}
	// group and team descriptors are managed by as the same object in azure devops APIs
	groupAndTeamDescriptors, err := resolvers.ResolveGroupAndTeamDescriptors(ctx, e.kube, cr.Spec.GroupsRefs, cr.Spec.TeamsRefs)
	if err != nil {
		return err
	}
	var res *groups.GroupResponse
	if cr.Spec.OriginID != nil {
		// Group is an Azure Active Directory user
		res, err = groups.Create(ctx, e.azCli, groups.CreateOptions[groups.SetGroupMembership]{
			Organization: helpers.String(organization),
			GroupData: groups.SetGroupMembership{
				OriginID: *cr.Spec.OriginID,
			},
			ScopeDescriptor: scopeDescriptor,
		})
		if err != nil {
			return fmt.Errorf("failed to add AAD group: %w", err)
		}
	} else {
		res, err = groups.Create(ctx, e.azCli, groups.CreateOptions[groups.GroupDescription]{
			Organization:    helpers.String(organization),
			ScopeDescriptor: scopeDescriptor,
			GroupData: groups.GroupDescription{
				DisplayName: helpers.String(cr.Spec.GroupsName),
				Description: cr.Spec.Description,
			}})
		if err != nil {
			return fmt.Errorf("failed to create group: %w", err)
		}
	}

	for _, containerDescriptor := range groupAndTeamDescriptors {
		err = memberships.Create(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   res.Descriptor,
			ContainerDescriptor: containerDescriptor,
		})
		if err != nil {
			return fmt.Errorf("failed to create group membership: %w", err)
		}
	}

	cr.Status.Descriptor = helpers.StringPtr(res.Descriptor)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*groupsv1alpha1.Groups)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	_, organization, err := resolveProjectAndOrganization(ctx, e.kube, cr)
	if err != nil {
		return err
	}

	err = groups.Delete(ctx, e.azCli, groups.DeleteOptions{
		Organization:    helpers.String(organization),
		GroupDescriptor: helpers.String(cr.Status.Descriptor),
	})
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return nil
}
