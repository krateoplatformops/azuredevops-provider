package feeds

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	feedsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/feeds/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/feeds"
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
	errNotCR = "managed resource is not a Feed custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(feedsv1alpha1.FeedGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(feedsv1alpha1.FeedGroupVersionKind),
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
		For(&feedsv1alpha1.Feed{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*feedsv1alpha1.Feed)
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
	cr, ok := mg.(*feedsv1alpha1.Feed)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var observed *feeds.Feed
	if cr.Status.Id == nil {
		observed, err = e.findFeed(ctx, cr)
	} else {
		observed, err = feeds.Get(ctx, e.azCli, feeds.GetOptions{
			Organization: organization,
			Project:      project,
			FeedId:       helpers.String(cr.Status.Id),
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

	cr.Status.Id = observed.Id
	cr.Status.Url = observed.Url

	err = e.kube.Status().Update(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	if !compareFeed(ctx, cr, observed) {
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
		}, nil
	}

	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*feedsv1alpha1.Feed)
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

	var upstreams []feeds.UpstreamSource
	for _, v := range cr.Spec.UpstreamSources {
		upstreams = append(upstreams, feeds.UpstreamSource{
			Name:               v.Name,
			Location:           v.Location,
			Protocol:           v.Protocol,
			UpstreamSourceType: v.UpstreamSourceType,
		})
	}

	res, err := feeds.Create(ctx, e.azCli, feeds.CreateOptions{
		Organization: organization,
		Project:      project,
		Feed: &feeds.Feed{
			Name:            name,
			IsReadOnly:      helpers.BoolOrDefault(cr.Spec.IsReadOnly, false),
			UpstreamSources: upstreams,
		},
	})
	if err != nil {
		return err
	}

	cr.Status.Id = helpers.StringPtr(*res.Id)
	cr.Status.Url = helpers.StringPtr(*res.Url)

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {

	cr, ok := mg.(*feedsv1alpha1.Feed)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return err
	}

	feed, err := e.localFeedUpdate(ctx, cr, organization, project)
	if err != nil {
		return err
	}

	_, err = feeds.Update(ctx, e.azCli, feeds.UpdateOptions{
		Organization: organization,
		Project:      project,
		FeedId:       helpers.String(cr.Status.Id),
		FeedUpdate: &feeds.FeedUpdate{
			Name:                       feed.Name,
			Description:                feed.Description,
			UpstreamEnabled:            feed.UpstreamEnabled,
			UpstreamSources:            feed.UpstreamSources,
			HideDeletedPackageVersions: feed.HideDeletedPackageVersions,
			DefaultViewId:              feed.DefaultViewId,
			BadgesEnabled:              feed.BadgesEnabled,
		},
	})

	return err
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*feedsv1alpha1.Feed)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	feedId := helpers.String(cr.Status.Id)
	if len(feedId) == 0 {
		return fmt.Errorf("missing Feed identifier")
	}

	organization, project, err := e.resolveProjectAndOrg(ctx, cr.DeepCopy())
	if err != nil {
		return err
	}

	cr.SetConditions(rtv1.Deleting())

	return feeds.Delete(ctx, e.azCli, feeds.DeleteOptions{
		Organization: organization,
		Project:      project,
		FeedId:       feedId,
	})
}

func (e *external) resolveProjectAndOrg(ctx context.Context, cr *feedsv1alpha1.Feed) (string, string, error) {
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

func (e *external) findFeed(ctx context.Context, cr *feedsv1alpha1.Feed) (*feeds.Feed, error) {
	org, prj, err := e.resolveProjectAndOrg(ctx, cr)
	if err != nil {
		return nil, err
	}

	name := helpers.String(cr.Spec.Name)
	if len(name) == 0 {
		name = cr.GetName()
	}

	return feeds.Find(ctx, e.azCli, feeds.FindOptions{
		Organization: org,
		Project:      prj,
		FeedName:     name,
	})
}
func findUpstream(upstream feeds.UpstreamSource, upstreams []feeds.UpstreamSource) bool {
	for _, v := range upstreams {
		if helpers.String(upstream.Name) == helpers.String(v.Name) &&
			helpers.String(upstream.Location) == helpers.String(v.DisplayLocation) &&
			strings.EqualFold(helpers.String(upstream.Protocol), helpers.String(v.Protocol)) &&
			strings.EqualFold(helpers.String(upstream.UpstreamSourceType), helpers.String(v.UpstreamSourceType)) {
			return true
		}
	}
	return false
}
func compareFeed(ctx context.Context, cr *feedsv1alpha1.Feed, feed *feeds.Feed) bool {
	name := helpers.StringOrDefault(cr.Spec.Name, cr.GetName())

	if name != feed.Name {
		return false
	}
	for _, e := range cr.Spec.UpstreamSources {
		upstream := feeds.UpstreamSource{
			Name:               e.Name,
			Location:           e.Location,
			Protocol:           e.Protocol,
			UpstreamSourceType: e.UpstreamSourceType,
		}
		if !findUpstream(upstream, feed.UpstreamSources) {
			return false
		}
	}

	return true
}

func (e *external) localFeedUpdate(ctx context.Context, cr *feedsv1alpha1.Feed, organization string, project string) (*feeds.Feed, error) {
	feed, err := feeds.Get(ctx, e.azCli, feeds.GetOptions{
		Organization: organization,
		Project:      project,
		FeedId:       helpers.String(cr.Status.Id),
	})
	if err != nil {
		return nil, err
	}
	if feed == nil {
		return nil, fmt.Errorf("feed not found")
	}

	feed.Name = helpers.StringOrDefault(cr.Spec.Name, feed.Name)
	feed.IsReadOnly = helpers.BoolOrDefault(cr.Spec.IsReadOnly, feed.IsReadOnly)
	for _, v := range cr.Spec.UpstreamSources {
		f := feeds.UpstreamSource{
			Name:               v.Name,
			Location:           v.Location,
			Protocol:           v.Protocol,
			UpstreamSourceType: v.UpstreamSourceType,
		}
		if !findUpstream(f, feed.UpstreamSources) {
			feed.UpstreamSources = append(feed.UpstreamSources, f)
		}

	}
	return feed, nil
}
