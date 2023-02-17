package teamproject

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httplib"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
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

	teamprojectv1alpha1 "gihtub.com/krateoplatformops/azuredevops-provider/apis/teamproject/v1alpha1"
)

const (
	errNotTeamProject = "managed resource is not a TeamProject custom resource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(teamprojectv1alpha1.TeamProjectGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(teamprojectv1alpha1.TeamProjectGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(recorder)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&teamprojectv1alpha1.TeamProject{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*teamprojectv1alpha1.TeamProject)
	if !ok {
		return nil, errors.New(errNotTeamProject)
	}

	spec := cr.Spec.DeepCopy()

	csr := spec.Credentials.SecretRef
	if csr == nil {
		return nil, fmt.Errorf("no credentials secret referenced")
	}

	token, err := resource.GetSecret(ctx, c.kube, csr.DeepCopy())
	if err != nil {
		return nil, err
	}

	opts := azuredevops.Options{
		BaseURL: spec.ApiUrl,
		Verbose: helpers.IsBoolPtrEqualToBool(spec.Verbose, true),
		Token:   token,
	}

	httpClient := httplib.CreateHTTPClient(httplib.CreateHTTPClientOpts{
		Timeout: 40 * time.Second,
	})

	return &external{
		kube:  c.kube,
		log:   c.log,
		azCli: azuredevops.NewClient(httpClient, opts),
		rec:   c.recorder,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	kube  client.Client
	log   logging.Logger
	azCli *azuredevops.Client
	rec   record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*teamprojectv1alpha1.TeamProject)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTeamProject)
	}

	if meta.GetExternalOperation(cr) != "" {
		op, err := azuredevops.GetOperation(ctx, e.azCli, azuredevops.GetOperationOpts{
			Organization: cr.Spec.Org,
			OperationId:  meta.GetExternalOperation(cr),
		})
		if err != nil {
			return managed.ExternalObservation{}, err
		}

		if op.Status != azuredevops.StatusSucceeded {
			return managed.ExternalObservation{}, nil
		}

		prj, err := findTeamProject(ctx, e.azCli, cr.Spec.Org, cr.Spec.Name)
		if err != nil {
			return managed.ExternalObservation{}, err
		}

		e.log.Debug("Found Project", "id", *prj.Id, "name", *prj.Name)

		meta.RemoveAnnotations(cr, meta.AnnotationKeyExternalOperation)
		meta.SetExternalName(cr, helpers.String(prj.Id))

		cr.Status.Id = prj.Id
		cr.Status.Name = prj.Name
		cr.Status.Revision = prj.Revision
		cr.Status.State = (*string)(prj.State)

		cr.SetConditions(rtv1.Available())

		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*teamprojectv1alpha1.TeamProject)
	if !ok {
		return errors.New(errNotTeamProject)
	}

	if meta.GetExternalOperation(cr) != "" {
		return nil
	}

	cr.SetConditions(rtv1.Creating())

	spec := cr.Spec.DeepCopy()

	op, err := azuredevops.CreateProject(ctx, e.azCli, azuredevops.CreateProjectOpts{
		Organization: spec.Org,
		TeamProject:  teamProjectFromSpec(spec),
	})
	if err != nil {
		return err
	}

	meta.SetExternalOperation(cr, op.Id)
	cr.SetConditions(conditionFromOperationReference(op))

	e.log.Debug("Creating TeamProject", "org", spec.Org, "name", spec.Name, "status", op.Status)
	//e.rec.Eventf(cr, corev1.EventTypeNormal, "TeamProjectCreated", "TeamProject '%s/%s' created", spec.Org, spec.Name)

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*teamprojectv1alpha1.TeamProject)
	if !ok {
		return errors.New(errNotTeamProject)
	}

	cr.SetConditions(rtv1.Deleting())

	_, err := azuredevops.DeleteProject(ctx, e.azCli, azuredevops.DeleteProjectOpts{
		Organization: cr.Spec.Org,
		ProjectId:    helpers.String(cr.Status.Id),
	})
	if err != nil {
		return err
	}

	e.log.Debug("Delete TeamProject",
		"id", helpers.String(cr.Status.Id), "org", cr.Spec.Org, "name", cr.Spec.Name)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "TeamProjectDeleted",
		"TeamProject '%s/%s' deleted", cr.Spec.Org, cr.Spec.Name)

	return nil
}
