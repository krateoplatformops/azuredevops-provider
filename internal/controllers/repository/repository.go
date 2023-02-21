package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
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
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler/managed"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	repositoryv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/repository/v1alpha1"
)

const (
	errNotGitRepository = "managed resource is not a GitRepository custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(repositoryv1alpha1.GitRepositoryGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(repositoryv1alpha1.GitRepositoryGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(log),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithRecorder(event.NewAPIRecorder(recorder)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&repositoryv1alpha1.GitRepository{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*repositoryv1alpha1.GitRepository)
	if !ok {
		return nil, errors.New(errNotGitRepository)
	}

	cfg := cr.Spec.ConnectorConfig

	csr := cfg.Credentials.SecretRef
	if csr == nil {
		return nil, fmt.Errorf("no credentials secret referenced")
	}

	token, err := resource.GetSecret(ctx, c.kube, csr.DeepCopy())
	if err != nil {
		return nil, err
	}

	return &external{
		kube: c.kube,
		log:  c.log,
		azCli: azuredevops.NewClient(azuredevops.ClientOptions{
			BaseURL: cfg.ApiUrl,
			Verbose: helpers.IsBoolPtrEqualToBool(cfg.Verbose, true),
			Token:   token,
		}),
		rec: c.recorder,
	}, nil
}

type external struct {
	kube  client.Client
	log   logging.Logger
	azCli *azuredevops.Client
	rec   record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*repositoryv1alpha1.GitRepository)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGitRepository)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	spec := cr.Spec.DeepCopy()

	res, err := e.azCli.GetRepository(ctx, azuredevops.GetRepositoryOptions{
		Organization: spec.Organization,
		Project:      spec.Project,
		Repository:   meta.GetExternalName(cr),
	})
	if err != nil {
		return managed.ExternalObservation{}, resource.Ignore(httplib.IsNotFoundError, err)
	}
	if res == nil {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	meta.SetExternalName(cr, helpers.String(res.Id))

	cr.Status.Id = helpers.String(res.Id)
	cr.Status.DefaultBranch = helpers.String(res.DefaultBranch)
	cr.Status.RemoteUrl = helpers.String(res.RemoteUrl)
	cr.Status.SshUrl = helpers.String(res.SshUrl)
	cr.Status.Url = helpers.String(res.Url)

	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Update(ctx, cr)

}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*repositoryv1alpha1.GitRepository)
	if !ok {
		return errors.New(errNotGitRepository)
	}

	cr.SetConditions(rtv1.Creating())

	res, err := e.azCli.CreateRepository(ctx, azuredevops.CreateRepositoryOptions{
		Organization: cr.Spec.Organization,
		ProjectId:    cr.Spec.Project,
		Name:         cr.Spec.Name,
	})
	if err != nil {
		return err
	}

	if res != nil {
		e.log.Debug("GitRepository created", "id", helpers.String(res.Id), "url", helpers.String(res.Url))
		e.rec.Eventf(cr, corev1.EventTypeNormal, "GitRepositoryCreated",
			"GitRepository '%s' created", helpers.String(res.Url))
	}

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*repositoryv1alpha1.GitRepository)
	if !ok {
		return errors.New(errNotGitRepository)
	}

	spec := cr.Spec.DeepCopy()
	status := cr.Status.DeepCopy()

	cr.SetConditions(rtv1.Deleting())

	err := e.azCli.DeleteRepository(ctx, azuredevops.DeleteRepositoryOptions{
		Organization: spec.Organization,
		Project:      spec.Project,
		RepositoryId: status.Id,
	})
	if err != nil {
		return resource.Ignore(httplib.IsNotFoundError, err)
	}

	e.log.Debug("GitRepository deleted", "id", status.Id, "url", status.Url)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "GitRepositoryDeleted",
		"GitRepository '%s' deleted", status.Url)

	return nil
}
