package pullrequests

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
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/pullrequests"

	pullrequestsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/pullrequests/v1alpha1"
)

const (
	errNotPullRequest = "managed resource is not a PullRequest custom resource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(pullrequestsv1alpha1.PullRequestGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(pullrequestsv1alpha1.PullRequestGroupVersionKind),
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
		For(&pullrequestsv1alpha1.PullRequest{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*pullrequestsv1alpha1.PullRequest)
	if !ok {
		return nil, errors.New(errNotPullRequest)
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
	cr, ok := mg.(*pullrequestsv1alpha1.PullRequest)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotPullRequest)
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, fmt.Errorf("failed to resolve project reference: %v", err)
	}
	repository, err := resolvers.ResolveGitRepository(ctx, e.kube, cr.Spec.RepositoryRef)
	if err != nil {
		return reconciler.ExternalObservation{}, fmt.Errorf("failed to resolve repository reference: %v", err)
	}

	response := &pullrequests.PullRequest{}

	if cr.Status.Id == nil {
		response, err = pullrequests.Find(ctx, e.azCli, pullrequests.FindOptions{
			Organization:  project.Spec.Organization,
			ProjectId:     project.Status.Id,
			RepositoryId:  repository.Status.Id,
			Title:         cr.Spec.PullRequest.Title,
			SourceRefName: cr.Spec.PullRequest.SourceRefName,
			TargetRefName: cr.Spec.PullRequest.TargetRefName,
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		response, err = pullrequests.Get(ctx, e.azCli, pullrequests.GetOptions{
			Organization:  project.Spec.Organization,
			ProjectId:     project.Status.Id,
			RepositoryId:  repository.Status.Id,
			PullRequestId: helpers.String(cr.Status.Id),
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	}
	if response == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	cr.Status.Id = helpers.StringPtr(fmt.Sprintf("%d", response.PullRequestId))
	cr.SetConditions(rtv1.Available())

	err = e.kube.Status().Update(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	if !isUpdated(cr, response) {
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
	cr, ok := mg.(*pullrequestsv1alpha1.PullRequest)
	if !ok {
		return errors.New(errNotPullRequest)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	e.log.Info("Creating resource")

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve project reference: %v", err)
	}

	repo, err := resolvers.ResolveGitRepository(ctx, e.kube, cr.Spec.RepositoryRef)
	if err != nil {
		return fmt.Errorf("failed to resolve repository reference: %v", err)
	}

	pr, err := customResourceToPullRequest(cr)
	if err != nil {
		return fmt.Errorf("failed to convert custom resource to pull request: %v", err)
	}

	response, err := pullrequests.Create(ctx, e.azCli, pullrequests.CreateOptions{
		PullRequest:  pr,
		Organization: project.Spec.Organization,
		ProjectId:    project.Status.Id,
		RepositoryId: repo.Status.Id,
	})

	if err != nil {
		return fmt.Errorf("failed to create pull request: %v", err)
	}

	cr.SetConditions(rtv1.Creating())
	cr.Status.Id = helpers.StringPtr(fmt.Sprintf("%d", response.PullRequestId))

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*pullrequestsv1alpha1.PullRequest)
	if !ok {
		return errors.New(errNotPullRequest)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	e.log.Info("Updating resource")

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve project reference: %v", err)
	}
	repository, err := resolvers.ResolveGitRepository(ctx, e.kube, cr.Spec.RepositoryRef)
	if err != nil {
		return fmt.Errorf("failed to resolve repository reference: %v", err)
	}
	getResponse, err := pullrequests.Get(ctx, e.azCli, pullrequests.GetOptions{
		Organization:  project.Spec.Organization,
		ProjectId:     project.Status.Id,
		RepositoryId:  repository.Status.Id,
		PullRequestId: helpers.String(cr.Status.Id),
	})
	if err != nil && !httplib.IsNotFoundError(err) {
		return fmt.Errorf("failed to get pull request: %v", err)
	}

	pr, err := createPRWithModifiedFields(cr, getResponse)
	if err != nil {
		return fmt.Errorf("failed to create pull request with modified fields: %v", err)
	}

	response, err := pullrequests.Update(ctx, e.azCli, pullrequests.UpdateOptions{
		PullRequest:   pr,
		Organization:  project.Spec.Organization,
		ProjectId:     project.Status.Id,
		RepositoryId:  repository.Status.Id,
		PullRequestId: helpers.String(cr.Status.Id),
	})

	if err != nil {
		return fmt.Errorf("failed to update pull request: %v", err)
	}

	cr.SetConditions(rtv1.Creating())
	cr.Status.Id = helpers.StringPtr(fmt.Sprintf("%d", response.PullRequestId))

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // The pull request should not be deleted - a "delete" action is not supported - set the status to "abandoned" instead
}
