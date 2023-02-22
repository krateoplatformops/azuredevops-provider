package pipeline

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/event"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler/managed"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"
	"github.com/lucasepe/httplib"
	"github.com/pkg/errors"

	connectorconfigs "github.com/krateoplatformops/azuredevops-provider/apis/connectorconfigs/v1alpha1"
	pipelines "github.com/krateoplatformops/azuredevops-provider/apis/pipelines/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
)

const (
	errNotPipeline                = "managed resource is not a Pipeline custom resource"
	annotationKeyConnectorVerbose = "krateo.io/connector-verbose"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(pipelines.PipelineGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(pipelines.PipelineGroupVersionKind),
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
		For(&pipelines.Pipeline{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*pipelines.Pipeline)
	if !ok {
		return nil, errors.New(errNotPipeline)
	}

	opts, err := c.clientOptions(ctx, cr.Spec.ConnectorConfigRef)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(cr.GetAnnotations()[annotationKeyConnectorVerbose], "true") {
		opts.Verbose = true
	}

	return &external{
		kube:  c.kube,
		log:   c.log,
		azCli: azuredevops.NewClient(opts),
		rec:   c.recorder,
	}, nil
}

func (c *connector) clientOptions(ctx context.Context, ref *pipelines.Selector) (azuredevops.ClientOptions, error) {
	opts := azuredevops.ClientOptions{}

	if ref == nil {
		return opts, errors.New("no ConnectorConfig referenced")
	}

	cfg := connectorconfigs.ConnectorConfig{}
	err := c.kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &cfg)
	if err != nil {
		return opts, errors.Wrapf(err, "cannot get %s connector config", ref.Name)
	}

	csr := cfg.Spec.Credentials.SecretRef
	if csr == nil {
		return opts, fmt.Errorf("no credentials secret referenced")
	}

	sec := corev1.Secret{}
	err = c.kube.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, &sec)
	if err != nil {
		return opts, errors.Wrapf(err, "cannot get %s secret", ref.Name)
	}

	token, err := resource.GetSecret(ctx, c.kube, csr.DeepCopy())
	if err != nil {
		return opts, err
	}

	opts.BaseURL = cfg.Spec.ApiUrl
	opts.Token = token
	opts.Verbose = false

	return opts, nil
}

type external struct {
	kube  client.Client
	log   logging.Logger
	azCli *azuredevops.Client
	rec   record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*pipelines.Pipeline)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPipeline)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	spec := cr.Spec.DeepCopy()

	res, err := e.azCli.GetPipeline(ctx, azuredevops.GetPipelineOptions{
		Organization: spec.Organization,
		Project:      helpers.String(spec.Project),
		PipelineId:   meta.GetExternalName(cr),
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

	meta.SetExternalName(cr, fmt.Sprintf("%d", *res.Id))
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.Id = helpers.StringPtr(fmt.Sprintf("%d", *res.Id))
	cr.Status.Url = helpers.StringPtr(*res.Url)

	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil

}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*pipelines.Pipeline)
	if !ok {
		return errors.New(errNotPipeline)
	}

	cr.SetConditions(rtv1.Creating())

	spec := cr.Spec.DeepCopy()

	res, err := e.azCli.CreatePipeline(ctx, azuredevops.CreatePipelineOptions{
		Organization: spec.Organization,
		Project:      helpers.String(spec.Project),
		Pipeline: azuredevops.Pipeline{
			Folder: spec.Folder,
			Name:   spec.Name,
			Configuration: &azuredevops.PipelineConfiguration{
				Type: azuredevops.ConfigurationType(*spec.ConfigurationType),
				Path: spec.DefinitionPath,
				Repository: &azuredevops.BuildRepository{
					Id:   repo.Id,
					Name: repo.Name,
					Type: azuredevops.BuildRepositoryType(*spec.RepositoryType),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	pipelineId := fmt.Sprintf("%d", *res.Id)
	meta.SetExternalName(cr, pipelineId)
	if err := e.kube.Update(ctx, cr); err != nil {
		return err
	}

	e.log.Debug("Pipeline created", "id", pipelineId, "url", helpers.String(res.Url))
	e.rec.Eventf(cr, corev1.EventTypeNormal, "GitRepositoryCreated",
		"Pipeline '%s' created", helpers.String(res.Url))

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}
