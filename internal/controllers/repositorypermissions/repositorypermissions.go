package repositorypermissions

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	repositorypermissionsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/repositorypermissions/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/identities"
	repositoryspermissions "github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/repositorypermissions"
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
	errNotRepository       = "managed resource is not a RepositoryPermission custom resource"
	errUnspecifiedResource = "repository permissions resource is not specified"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(repositorypermissionsv1alpha1.RepositoryPermissionGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(repositorypermissionsv1alpha1.RepositoryPermissionGroupVersionKind),
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
		For(&repositorypermissionsv1alpha1.RepositoryPermission{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*repositorypermissionsv1alpha1.RepositoryPermission)
	if !ok {
		return nil, errors.New(errNotRepository)
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
	cr, ok := mg.(*repositorypermissionsv1alpha1.RepositoryPermission)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotRepository)
	}

	repository, err := resolvers.ResolveGitRepository(ctx, e.kube, cr.Spec.RepositoryRef)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, repository.Spec.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	projectIdentity, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.Permissions.Identity.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	identityResponse, err := identities.Get(ctx, e.azCli, identities.GetOptions{
		Organization: projectIdentity.Spec.Organization,
		IdentityParams: identities.IdentityParams{
			Type:    identities.UserType(helpers.String(cr.Spec.Permissions.Identity.Type)),
			Name:    helpers.String(cr.Spec.Permissions.Identity.Name),
			Project: projectIdentity,
		},
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	identityDescriptor, err := identityResponse.IdentityMatch(&identities.IdentityParams{
		Type:    identities.UserType(helpers.String(cr.Spec.Permissions.Identity.Type)),
		Name:    helpers.String(cr.Spec.Permissions.Identity.Name),
		Project: projectIdentity,
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	res, err := repositoryspermissions.Get(ctx, e.azCli, repositoryspermissions.GetOptions{
		Organization: project.Spec.Organization,
		Token:        repositoryspermissions.CreateToken(project.Status.Id, repository.Status.Id),
		Descriptor:   identityDescriptor.Descriptor,
	})
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}
	cr.SetConditions(rtv1.Available())

	updateReconciler := reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: false,
	}

	if res.Count == 0 {
		return updateReconciler, nil
	}

	allowBits := resolvePermissionBits(cr.Spec.Permissions.AllowList)
	denyBits := resolvePermissionBits(cr.Spec.Permissions.DenyList)

	// If the merge flag is false, we need to check if the permission bits are the same as the ones we want to set.
	if !cr.Spec.Permissions.Merge && (res.Value[0].Allow != allowBits || res.Value[0].Deny != denyBits) {
		return updateReconciler, nil
	}

	// If the merge flag is true, we need to check if the permission bit setted on the spec are the same as the ones we want to set.
	if !comparePermissionBits(res.Value[0].Allow, allowBits) || !comparePermissionBits(res.Value[0].Deny, denyBits) {
		return updateReconciler, nil
	}

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	return nil // NOOP
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*repositorypermissionsv1alpha1.RepositoryPermission)
	if !ok {
		return errors.New(errNotRepository)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	spec := cr.Spec.DeepCopy()

	repository, err := resolvers.ResolveGitRepository(ctx, e.kube, cr.Spec.RepositoryRef)
	if err != nil {
		return err
	}
	projectRepo, err := resolvers.ResolveTeamProject(ctx, e.kube, repository.Spec.ProjectRef)
	if err != nil {
		return err
	}
	projectIdentity, err := resolvers.ResolveTeamProject(ctx, e.kube, spec.Permissions.Identity.ProjectRef)
	if err != nil {
		return err
	}

	identityResponse, err := identities.Get(ctx, e.azCli, identities.GetOptions{
		Organization: projectIdentity.Spec.Organization,
		IdentityParams: identities.IdentityParams{
			Type:    identities.UserType(helpers.String(spec.Permissions.Identity.Type)),
			Name:    helpers.String(spec.Permissions.Identity.Name),
			Project: projectIdentity,
		},
	})
	if err != nil {
		return err
	}

	identityDescriptor, err := identityResponse.IdentityMatch(&identities.IdentityParams{
		Type:    identities.UserType(helpers.String(spec.Permissions.Identity.Type)),
		Name:    helpers.String(spec.Permissions.Identity.Name),
		Project: projectIdentity,
	})
	if err != nil {
		return err
	}

	updateResponse, err := repositoryspermissions.Update(ctx, e.azCli, repositoryspermissions.UpdateOptions{
		Organization: projectRepo.Spec.Organization,
		ResourceAuthorization: &repositoryspermissions.AccessControlUpdate{
			Merge: spec.Permissions.Merge,
			Token: repositoryspermissions.CreateToken(projectRepo.Status.Id, repository.Status.Id),
			AccessControlEntries: []repositoryspermissions.AccessControlEntry{
				{
					Descriptor: identityDescriptor.Descriptor,
					Allow:      resolvePermissionBits(spec.Permissions.AllowList),
					Deny:       resolvePermissionBits(spec.Permissions.DenyList),
				},
			},
		}})
	if err != nil {
		e.rec.Eventf(cr, corev1.EventTypeWarning, "RepositoryPermissionUpdateFailed",
			"Failed Update to Repository '%s' with id: %s with error: %s", repository.Name, repository.Status.Id, err.Error())
		return err
	}

	cr.Status.IdentityDescriptor = identityDescriptor.Descriptor
	cr.Status.AllowPermissionBit = helpers.IntPtr(updateResponse.Value[0].Allow)
	cr.Status.DenyPermissionBit = helpers.IntPtr(updateResponse.Value[0].Deny)

	e.log.Debug("RepositoryPermission updated", "Repository id", repository.Status.Id,
		"Repository name", repository.Name)

	e.rec.Eventf(cr, corev1.EventTypeNormal, "RepositoryPermissionUpdated",
		"Repository Permission of repo '%s' updated", repository.Status.Id)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}
func resolvePermissionBits(perms []string) int {
	retPerm := 0
	for _, perm := range perms {
		retPerm = retPerm + repositoryspermissions.PermissionBitValue(perm)
	}
	return retPerm
}

func comparePermissionBits(setted int, check int) bool {
	mask := setted & check
	return mask == check
}
