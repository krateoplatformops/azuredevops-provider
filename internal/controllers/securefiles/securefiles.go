package securefiles

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securefilesv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/securefiles/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/securefiles"
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
	errNotCR = "managed resource is not a SecureFiles custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(securefilesv1alpha1.SecureFilesGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(securefilesv1alpha1.SecureFilesGroupVersionKind),
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
		For(&securefilesv1alpha1.SecureFiles{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*securefilesv1alpha1.SecureFiles)
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
	cr, ok := mg.(*securefilesv1alpha1.SecureFiles)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	if cr.GetDeletionTimestamp() != nil && !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("Deletion of external resource is not allowed, skipping observation and deleting CR.")
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, fmt.Errorf("cannot resolve project reference: %w", err)
	}

	var observed *securefiles.SecureFileResource
	if cr.Status.Id != nil {
		observed, err = securefiles.Get(ctx, e.azCli, securefiles.GetOptions{
			Organization: project.Spec.Organization,
			Project:      project.Status.Id,
			SecretFileId: helpers.String(cr.Status.Id),
		})
		if httplib.IsNotFoundError(err) && cr.GetDeletionTimestamp() != nil {
			return reconciler.ExternalObservation{}, nil
		}
		if err != nil {
			return reconciler.ExternalObservation{}, fmt.Errorf("cannot get secure file: %w", err)
		}
	} else {
		observed, err = securefiles.Find(ctx, e.azCli, securefiles.FindOptions{
			ListOptions: securefiles.ListOptions{
				Organization: project.Spec.Organization,
				Project:      project.Status.Id,
			},
			SecureFileName: cr.Spec.Name,
		})
		if httplib.IsNotFoundError(err) && cr.GetDeletionTimestamp() != nil {
			return reconciler.ExternalObservation{}, nil
		}
		if err != nil {
			return reconciler.ExternalObservation{}, fmt.Errorf("cannot find secure file: %w", err)
		}
	}

	cr.Status.Id = &observed.ID

	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	return nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*securefilesv1alpha1.SecureFiles)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		cr.SetConditions(rtv1.Deleting())
		return e.kube.Status().Update(ctx, cr)
	}

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return fmt.Errorf("cannot resolve project reference: %w", err)
	}

	err = securefiles.Delete(ctx, e.azCli, securefiles.DeleteOptions{
		Organization: project.Spec.Organization,
		Project:      project.Status.Id,
		SecureFileId: helpers.String(cr.Status.Id),
	})
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return e.kube.Status().Update(ctx, cr)
}
