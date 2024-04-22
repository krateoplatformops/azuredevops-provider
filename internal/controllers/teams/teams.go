package teams

import (
	"context"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	teamsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/teams/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/descriptors"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/memberships"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/teams"
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
	errNotCR = "managed resource is not a Team custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(teamsv1alpha1.TeamGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(teamsv1alpha1.TeamGroupVersionKind),
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
		For(&teamsv1alpha1.Team{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*teamsv1alpha1.Team)
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

func resolveProjectAndOrganization(ctx context.Context, kube client.Client, cr *teamsv1alpha1.Team) (*string, *string, error) {
	var projectId, organization *string
	project, err := resolvers.ResolveTeamProject(ctx, kube, cr.Spec.ProjectRef)
	if err != nil {
		return nil, nil, err
	}
	projectId = &project.Status.Id
	organization = &project.Spec.Organization
	return projectId, organization, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (reconciler.ExternalObservation, error) {
	cr, ok := mg.(*teamsv1alpha1.Team)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}
	projectId, organization, err := resolveProjectAndOrganization(ctx, e.kube, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var team *teams.TeamResponse
	if cr.Status.Id == nil {
		team, err = teams.FindTeamByName(ctx, e.azCli, teams.FindTeamByNameOptions{
			ListOptions: teams.ListOptions{
				Organization: helpers.String(organization),
				ProjectID:    helpers.String(projectId),
			},
			TeamName:  cr.Spec.Name,
			ProjectID: helpers.String(projectId),
		})
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		team, err = teams.Get(ctx, e.azCli, teams.GetOptions{
			Organization: helpers.String(organization),
			ProjectID:    helpers.String(projectId),
			TeamID:       helpers.String(cr.Status.Id),
		})
		if err != nil && !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}
	if team == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	check := true
	groupDescriptors, err := resolvers.ResolveGroupListDescriptors(ctx, e.kube, cr.Spec.GroupRefs)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	teamDescriptor, err := descriptors.GetDescriptor(ctx, e.azCli, descriptors.GetOptions{
		Organization: helpers.String(organization),
		ResourceID:   team.ID,
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	for _, groupDescriptor := range groupDescriptors {
		err = memberships.CheckMembership(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   helpers.String(teamDescriptor),
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

	cr.Status.Descriptor = teamDescriptor
	cr.Status.Id = helpers.StringPtr(team.ID)
	cr.SetConditions(rtv1.Available())
	err = e.kube.Status().Update(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	if len(groupDescriptors) != 0 && !check {
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

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*teamsv1alpha1.Team)
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

	team, err := teams.Get(ctx, e.azCli, teams.GetOptions{
		Organization: helpers.String(organization),
		ProjectID:    helpers.String(projectId),
		TeamID:       helpers.String(cr.Status.Id),
	})
	if err != nil && !azuredevops.IsNotFound(err) {
		return err
	}
	if team == nil {
		return errors.New("team not found")
	}

	groupDescriptors, err := resolvers.ResolveGroupListDescriptors(ctx, e.kube, cr.Spec.GroupRefs)
	if err != nil {
		return err
	}
	teamDescriptor, err := descriptors.GetDescriptor(ctx, e.azCli, descriptors.GetOptions{
		Organization: helpers.String(organization),
		ResourceID:   team.ID,
	})
	if err != nil {
		return err
	}

	for _, groupDescriptor := range groupDescriptors {
		err = memberships.CheckMembership(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   helpers.String(teamDescriptor),
			ContainerDescriptor: groupDescriptor,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return err
		}
		err = memberships.Create(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   helpers.String(teamDescriptor),
			ContainerDescriptor: groupDescriptor,
		})
		if err != nil {
			return err
		}
	}

	cr.Status.Id = helpers.StringPtr(team.ID)
	cr.Status.Descriptor = teamDescriptor
	return e.kube.Status().Update(ctx, cr)
}
func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*teamsv1alpha1.Team)
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

	res, err := teams.Create(ctx, e.azCli, teams.CreateOptions{
		Organization: helpers.String(organization),
		ProjectID:    helpers.String(projectId),
		TeamData: teams.TeamData{
			Name:        cr.Spec.Name,
			Description: cr.Spec.Description,
		},
	})
	if err != nil {
		return err
	}
	teamDescriptor, err := descriptors.GetDescriptor(ctx, e.azCli, descriptors.GetOptions{
		Organization: helpers.String(organization),
		ResourceID:   res.ID,
	})
	if err != nil {
		return err
	}
	groupDescriptors, err := resolvers.ResolveGroupListDescriptors(ctx, e.kube, cr.Spec.GroupRefs)
	if err != nil {
		return err
	}

	for _, groupDescriptor := range groupDescriptors {
		err = memberships.Create(ctx, e.azCli, memberships.CheckMembershipOptions{
			Organization:        helpers.String(organization),
			SubjectDescriptor:   helpers.String(teamDescriptor),
			ContainerDescriptor: groupDescriptor,
		})
		if err != nil {
			return err
		}
	}

	cr.Status.Id = helpers.StringPtr(res.ID)
	cr.Status.Descriptor = teamDescriptor
	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*teamsv1alpha1.Team)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	projectId, organization, err := resolveProjectAndOrganization(ctx, e.kube, cr)
	if err != nil {
		return err
	}

	err = teams.Delete(ctx, e.azCli, teams.DeleteOptions{
		Organization: helpers.String(organization),
		ProjectID:    helpers.String(projectId),
		TeamID:       helpers.String(cr.Status.Id),
	})
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return nil
}
