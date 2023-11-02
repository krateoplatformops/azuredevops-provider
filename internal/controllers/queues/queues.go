package queues

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	queuesv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/queues/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/pools"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/queues"
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
	errNotCR = "managed resource is not a Queue custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(queuesv1alpha1.QueueGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(queuesv1alpha1.QueueGroupVersionKind),
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
		For(&queuesv1alpha1.Queue{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*queuesv1alpha1.Queue)
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
	cr, ok := mg.(*queuesv1alpha1.Queue)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var observed *queues.TaskAgentQueue
	if cr.Status.Id == nil {
		observed, err = e.findQueue(ctx, cr)
	} else {
		observed, err = queues.Get(ctx, e.azCli, queues.GetOptions{
			Organization: organization,
			Project:      project,
			QueueId:      helpers.Int(cr.Status.Id),
		})
	}
	if err != nil {
		if !azuredevops.IsNotFound(err) {
			return reconciler.ExternalObservation{}, err
		}
	}

	if observed == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	cr.SetConditions(rtv1.Available())

	cr.Status.Id = helpers.IntPtr(*observed.Id)

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, e.kube.Status().Update(ctx, cr)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*queuesv1alpha1.Queue)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Creating())

	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	all, err := pools.Find(ctx, e.azCli, pools.FindOptions{
		Organization: organization,
		PoolName:     cr.Spec.Pool,
	})
	if err != nil {
		return err
	}
	if len(all) < 1 {
		return fmt.Errorf("pool '%s' not found", cr.Spec.Pool)
	}

	res, err := queues.Add(ctx, e.azCli, queues.AddOptions{
		Organization: organization,
		Project:      project,
		Queue: &queues.TaskAgentQueue{
			Name: name,
			Pool: &queues.TaskAgentPoolReference{
				Id: all[0].Id,
			},
		},
	})
	if err != nil {
		return err
	}

	cr.Status.Id = helpers.IntPtr(*res.Id)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // NOOP
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*queuesv1alpha1.Queue)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	if cr.Status.Id == nil {
		return fmt.Errorf("missing Queue identifier")
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr.DeepCopy())
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return queues.Delete(ctx, e.azCli, queues.DeleteOptions{
		Organization: organization,
		Project:      project,
		QueueId:      helpers.Int(cr.Status.Id),
	})
}

func (e *external) resolveProjectAndOrg(ctx context.Context, cr *queuesv1alpha1.Queue) (string, string, error) {
	organization := helpers.StringOrDefault(cr.Spec.Organization, "")
	project := helpers.StringOrDefault(cr.Spec.Project, "")

	if cr.Spec.ProjectRef != nil {
		prj, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
		if err != nil {
			return "", "", errors.Wrapf(err, "unble to resolve TeamProject: %s", cr.Spec.ProjectRef.Name)
		}
		if prj != nil {
			project = prj.Spec.Name
			organization = prj.Spec.Organization
		}
	}

	if len(project) == 0 {
		return "", "", fmt.Errorf("missing Project name")
	}

	if len(organization) == 0 {
		return "", "", fmt.Errorf("missing Organization name")
	}

	return organization, project, nil
}

func (e *external) findQueue(ctx context.Context, cr *queuesv1alpha1.Queue) (*queues.TaskAgentQueue, error) {
	org, prj, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return nil, err
	}

	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	all, err := queues.FindByNames(ctx, e.azCli, queues.FindByNamesOptions{
		Organization: org,
		Project:      prj,
		QueueNames:   []string{name},
	})
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("queue '%s' not found", name)
	}

	return &all[0], nil
}
