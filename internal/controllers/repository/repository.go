package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/repositories"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
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
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	repositoriesv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/repositories/v1alpha1"
)

const (
	errNotGitRepository = "managed resource is not a GitRepository custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(repositoriesv1alpha1.GitRepositoryGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(repositoriesv1alpha1.GitRepositoryGroupVersionKind),
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
		For(&repositoriesv1alpha1.GitRepository{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*repositoriesv1alpha1.GitRepository)
	if !ok {
		return nil, errors.New(errNotGitRepository)
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
	cr, ok := mg.(*repositoriesv1alpha1.GitRepository)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotGitRepository)
	}

	spec := cr.Spec.DeepCopy()

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, spec.PojectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, errors.Wrapf(err, "unble to resolve TeamProject: %s", spec.PojectRef.Name)
	}

	var repo *repositories.GitRepository
	if repoId := meta.GetExternalName(cr); repoId != "" {
		var err error
		repo, err = repositories.Get(ctx, e.azCli, repositories.GetOptions{
			Organization: prj.Spec.Organization,
			Project:      prj.Status.Id,
			Repository:   repoId,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if repo == nil {
		var err error
		if len(prj.Status.Id) == 0 {
			return reconciler.ExternalObservation{}, errors.New("TeamProject is not initialized")
		}
		repo, err = repositories.Find(ctx, e.azCli, repositories.FindOptions{
			Organization: prj.Spec.Organization,
			Project:      prj.Status.Id,
			Name:         cr.Spec.Name,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if repo == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	meta.SetExternalName(cr, helpers.String(repo.Id))
	if err := e.kube.Update(ctx, cr); err != nil {
		return reconciler.ExternalObservation{}, err
	}

	cr.Status.Id = helpers.String(repo.Id)
	cr.Status.DefaultBranch = helpers.String(repo.DefaultBranch)
	cr.Status.SshUrl = helpers.String(repo.SshUrl)
	cr.Status.Url = helpers.String(repo.Url)
	cr.Status.RemoteUrl = helpers.String(repo.RemoteUrl)

	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*repositoriesv1alpha1.GitRepository)
	if !ok {
		return errors.New(errNotGitRepository)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	cr.SetConditions(rtv1.Creating())

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PojectRef)
	if err != nil || prj == nil {
		return errors.Wrapf(err, "unble to resolve TeamProject: %s", cr.Spec.PojectRef.Name)
	}

	res, err := repositories.Create(ctx, e.azCli, repositories.CreateOptions{
		Organization: prj.Spec.Organization,
		ProjectId:    prj.Status.Id,
		Name:         cr.Spec.Name,
	})
	if err != nil {
		return err
	}

	if helpers.Bool(cr.Spec.Initialize) {
		//repoId := meta.GetExternalName(cr)
		//if res != nil {
		repoId := helpers.String(res.Id)
		//}
		defaultBranch := helpers.String(cr.Spec.DefaultBranch)
		if len(defaultBranch) == 0 {
			defaultBranch = "refs/heads/main"
		}

		_, err = repositories.CreatePush(ctx, e.azCli, repositories.GitPushOptions{
			Organization: prj.Spec.Organization,
			Project:      prj.Status.Id,
			RepositoryId: repoId,
			Push: &repositories.GitPush{
				RefUpdates: &[]repositories.GitRefUpdate{
					{
						Name:        helpers.StringPtr(defaultBranch),
						OldObjectId: helpers.StringPtr("0000000000000000000000000000000000000000"),
					},
				},
				Commits: &[]repositories.GitCommitRef{
					{
						Comment: helpers.StringPtr("Initial commit."),
						Changes: []repositories.GitChange{
							{
								ChangeType: repositories.ChangeTypeAdd,
								Item: map[string]string{
									"path": "/README.md",
								},
								NewContent: &repositories.ItemContent{
									Content:     fmt.Sprintf("# %s", helpers.String(res.Name)),
									ContentType: repositories.ContentTypeRawText,
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}

	meta.SetExternalName(cr, helpers.String(res.Id))
	if err := e.kube.Update(ctx, cr); err != nil {
		return err
	}

	e.log.Debug("GitRepository created", "id", helpers.String(res.Id), "url", helpers.String(res.Url))
	e.rec.Eventf(cr, corev1.EventTypeNormal, "GitRepositoryCreated",
		"GitRepository '%s' created", helpers.String(res.Url))

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*repositoriesv1alpha1.GitRepository)
	if !ok {
		return errors.New(errNotGitRepository)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	cr.SetConditions(rtv1.Deleting())

	prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.PojectRef)
	if err != nil {
		return errors.Wrapf(err, "unble to resolve TeamProject: %s", cr.Spec.PojectRef.Name)
	}

	err = repositories.Delete(ctx, e.azCli, repositories.DeleteOptions{
		Organization: prj.Spec.Organization,
		Project:      prj.Status.Id,
		RepositoryId: cr.Status.Id,
	})
	if err != nil {
		return resource.Ignore(httplib.IsNotFoundError, err)
	}

	e.log.Debug("GitRepository deleted", "id", cr.Status.Id, "url", cr.Status.Url)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "GitRepositoryDeleted",
		"GitRepository '%s' deleted", cr.Status.Url)

	return nil
}
