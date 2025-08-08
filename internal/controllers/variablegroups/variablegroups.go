package variablegroups

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	variablegroupsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/variablegroups/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	vgclient "github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/variablegroups"
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
	errNotCR = "managed resource is not a Users custom resource"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(variablegroupsv1alpha1.VariableGroupsGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(variablegroupsv1alpha1.VariableGroupsGroupVersionKind),
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
		For(&variablegroupsv1alpha1.VariableGroups{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*variablegroupsv1alpha1.VariableGroups)
	if !ok {
		return nil, errors.New(errNotCR)
	}

	opts, err := resolvers.ResolveConnectorConfig(ctx, c.kube, cr.Spec.ConnectorConfigRef)
	if err != nil {
		return nil, err
	}

	log := c.log.WithValues("name", cr.Name, "apiVersion", cr.APIVersion, "kind", cr.Kind)

	opts.Verbose = meta.IsVerbose(cr)

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
	cr, ok := mg.(*variablegroupsv1alpha1.VariableGroups)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}

	var observed *vgclient.VariableGroupResponse
	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.VariableGroupProjectReferences[0].ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, fmt.Errorf("failed to resolve project reference: %w", err)
	}

	if cr.Status.Id != "" {
		intId, err := strconv.Atoi(cr.Status.Id)
		if err != nil {
			return reconciler.ExternalObservation{}, fmt.Errorf("failed to convert id to int: %w", err)
		}
		observed, err = vgclient.Get(ctx, e.azCli, vgclient.GetOptions{
			Organization:    project.Spec.Organization,
			Project:         project.Status.Id,
			VariableGroupId: intId,
		})
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		observed, err = vgclient.Find(ctx, e.azCli, vgclient.FindOptions{
			ListOptions: vgclient.ListOptions{
				Organization: project.Spec.Organization,
				Project:      project.Status.Id,
			},
			VariableGroupName: helpers.String(cr.Spec.Name),
		})
		if err != nil {
			return reconciler.ExternalObservation{}, fmt.Errorf("failed to find variable group: %w", err)
		}
	}
	if observed == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}
	// Updating the status of the CR
	cr.Status.Id = fmt.Sprintf("%d", observed.ID)
	cr.SetConditions(rtv1.Available())
	err = e.kube.Status().Update(ctx, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	isSynced, err := isSynced(*cr, observed)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	if !isSynced {
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
	cr, ok := mg.(*variablegroupsv1alpha1.VariableGroups)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionUpdate) {
		e.log.Debug("External resource should not be updated by provider, skip updating.")
		return nil
	}

	e.log.Info("Updating resource")

	projects := []vgclient.VariableGroupProjectReference{}
	var organization string

	if len(cr.Spec.VariableGroupProjectReferences) > 1 {
		return errors.New("multiple projects are not supported")
	}

	for _, vgProjectRef := range cr.Spec.VariableGroupProjectReferences {
		project, err := resolvers.ResolveTeamProject(ctx, e.kube, vgProjectRef.ProjectRef)
		if err != nil {
			return fmt.Errorf("failed to resolve project reference: %w", err)
		}
		projects = append(projects, vgclient.VariableGroupProjectReference{
			Name: helpers.String(vgProjectRef.Name),
			ProjectReference: vgclient.ProjectReference{
				Name: project.Spec.Name,
				ID:   project.Status.Id,
			},
		})
		organization = project.Spec.Organization
	}
	variables := map[string]vgclient.Variable{}
	for k, v := range cr.Spec.Variables {
		variables[k] = vgclient.Variable{
			IsReadOnly: v.IsReadOnly,
			Value:      v.Value,
			IsSecret:   v.IsSecret,
		}
	}
	intId, err := strconv.Atoi(cr.Status.Id)
	if err != nil {
		return fmt.Errorf("failed to convert id to int: %w", err)
	}

	vgclient.Update(ctx, e.azCli, vgclient.UpdateOptions{
		Organization:    organization,
		Project:         projects[0].ProjectReference.Name,
		VariableGroupId: intId,
		VariableGroup: &vgclient.VariableGroupBody{
			Type:                           helpers.String(cr.Spec.Type),
			Variables:                      variables,
			Name:                           cr.Name,
			Description:                    helpers.String(cr.Spec.Description),
			VariableGroupProjectReferences: projects,
		},
	})

	return e.kube.Status().Update(ctx, cr)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*variablegroupsv1alpha1.VariableGroups)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	e.log.Info("Creating resource")

	projects := []vgclient.VariableGroupProjectReference{}
	var organization string

	if len(cr.Spec.VariableGroupProjectReferences) > 1 {
		return errors.New("multiple projects are not supported")
	}

	for _, vgProjectRef := range cr.Spec.VariableGroupProjectReferences {
		project, err := resolvers.ResolveTeamProject(ctx, e.kube, vgProjectRef.ProjectRef)
		if err != nil {
			return fmt.Errorf("failed to resolve project reference: %w", err)
		}
		projects = append(projects, vgclient.VariableGroupProjectReference{
			Name: helpers.String(vgProjectRef.Name),
			ProjectReference: vgclient.ProjectReference{
				Name: project.Spec.Name,
				ID:   project.Status.Id,
			},
		})
		organization = project.Spec.Organization
	}
	variables := map[string]vgclient.Variable{}
	for k, v := range cr.Spec.Variables {
		variables[k] = vgclient.Variable{
			IsReadOnly: v.IsReadOnly,
			Value:      v.Value,
			IsSecret:   v.IsSecret,
		}
	}

	response, err := vgclient.Create(ctx, e.azCli, vgclient.CreateOptions{
		Organization: organization,
		Project:      projects[0].ProjectReference.Name,
		VariableGroup: &vgclient.VariableGroupBody{
			Type:                           helpers.String(cr.Spec.Type),
			Variables:                      variables,
			Name:                           cr.Name,
			Description:                    helpers.String(cr.Spec.Description),
			VariableGroupProjectReferences: projects,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create variable group: %w", err)
	}

	cr.Status.Id = fmt.Sprintf("%d", response.ID)
	cr.SetConditions(rtv1.Creating())
	return e.kube.Status().Update(ctx, cr)
}
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*variablegroupsv1alpha1.VariableGroups)
	if !ok {
		return errors.New(errNotCR)
	}

	if !meta.IsActionAllowed(cr, meta.ActionDelete) {
		e.log.Debug("External resource should not be deleted by provider, skip deleting.")
		return nil
	}

	e.log.Info("Deleting resource")

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.VariableGroupProjectReferences[0].ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to resolve project reference: %w", err)
	}
	intId, err := strconv.Atoi(cr.Status.Id)
	if err != nil {
		return fmt.Errorf("failed to convert id to int: %w", err)
	}

	err = vgclient.Delete(ctx, e.azCli, vgclient.DeleteOptions{
		Organization:    project.Spec.Organization,
		ProjectID:       project.Status.Id,
		VariableGroupId: intId,
	})
	if err != nil {
		return fmt.Errorf("failed to delete variable group: %w", err)
	}

	cr.SetConditions(rtv1.Deleting())

	return nil
}

func isSynced(cr variablegroupsv1alpha1.VariableGroups, observed *vgclient.VariableGroupResponse) (bool, error) {
	for k, v := range cr.Spec.Variables {
		if !observed.Variables[k].IsSecret &&
			observed.Variables[k].Value != v.Value ||
			observed.Variables[k].IsReadOnly != v.IsReadOnly {
			return false, nil
		}
	}
	return true, nil
}
