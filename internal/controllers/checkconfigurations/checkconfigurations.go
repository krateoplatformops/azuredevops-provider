package checkconfigurations

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	checkconfigurations1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/checkconfigurations/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/checkconfiguration"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/groups"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/graphs/users"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
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
	errNotCR = "managed resource is not a CheckConfiguration custom resource"
)

type CheckConfigurationType string

const (
	CheckConfigurationTypeApproval     CheckConfigurationType = "Approval"
	CheckConfigurationTypeTaskCheck    CheckConfigurationType = "Task Check"
	CheckConfigurationTypeExtendsCheck CheckConfigurationType = "Extends Check"
)

func (t CheckConfigurationType) String() string {
	return string(t)
}

func (t CheckConfigurationType) GetIdFromType() string {
	switch t {
	case CheckConfigurationTypeApproval:
		return "8C6F20A7-A545-4486-9777-F762FAFE0D4D"
	case CheckConfigurationTypeTaskCheck:
		return "fe1de3ee-a436-41b4-bb20-f6eb4cb879a7"
	case CheckConfigurationTypeExtendsCheck:
		return "4020E66E-B0F3-47E1-BC88-48F3CC59B5F3"
	default:
		return ""
	}
}

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(checkconfigurations1alpha1.CheckConfigurationGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(checkconfigurations1alpha1.CheckConfigurationGroupVersionKind),
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
		For(&checkconfigurations1alpha1.CheckConfiguration{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	cr, ok := mg.(*checkconfigurations1alpha1.CheckConfiguration)
	if !ok {
		return nil, errors.New(errNotCR)
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
	cr, ok := mg.(*checkconfigurations1alpha1.CheckConfiguration)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotCR)
	}
	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return reconciler.ExternalObservation{}, err
	}

	var res *checkconfiguration.CheckConfiguration
	if cr.Status.ID != nil {
		res, err = checkconfiguration.Get(ctx, e.azCli, checkconfiguration.GetOptions{
			Organization: project.Spec.Organization,
			Project:      project.Status.Id,
			CheckID:      helpers.String(cr.Status.ID),
		})
		if err != nil && !httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{}, err
		}
	} else {
		resourceId, err := resolvers.ResolveResourceId(ctx, e.kube, cr.Spec.Resource.ResourceRef, cr.Spec.Resource.Type)
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
		res, err = checkconfiguration.Find(ctx, e.azCli, checkconfiguration.FindOptions{
			ListOptions: checkconfiguration.ListOptions{
				Organization: project.Spec.Organization,
				Project:      project.Status.Id,
				ResourceType: helpers.StringPtr(strings.ToLower(cr.Spec.Resource.Type)),
				ResourceId:   resourceId,
			},
			Type: checkconfiguration.Type{
				ID:   CheckConfigurationType(cr.Spec.Type).GetIdFromType(),
				Name: CheckConfigurationType(cr.Spec.Type).String(),
			},
		})
		if httplib.IsNotFoundError(err) {
			return reconciler.ExternalObservation{
				ResourceExists:   false,
				ResourceUpToDate: false,
			}, nil
		}
		if err != nil {
			return reconciler.ExternalObservation{}, err
		}
	}
	if res == nil {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: false,
		}, nil
	}

	cr.Status.ID = helpers.StringPtr(fmt.Sprintf("%v", res.ID))
	e.kube.Status().Update(ctx, cr)
	cr.SetConditions(rtv1.Available())

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*checkconfigurations1alpha1.CheckConfiguration)
	if !ok {
		return errors.New(errNotCR)
	}
	if !meta.IsActionAllowed(cr, meta.ActionCreate) {
		e.log.Debug("External resource should not be created by provider, skip creating.")
		return nil
	}

	e.log.Info("Creating resource")

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return err
	}
	resourceId, err := resolvers.ResolveResourceId(ctx, e.kube, cr.Spec.Resource.ResourceRef, cr.Spec.Resource.Type)
	if err != nil {
		return err
	}
	if resourceId == nil {
		return errors.New("resourceId is nil")
	}

	var approvers []checkconfiguration.Approver
	for _, approver := range cr.Spec.ApprovalSettings.Approvers {
		if approver.ID != nil {
			approvers = append(approvers, checkconfiguration.Approver{
				ID: helpers.String(approver.ID),
			})
		} else if approver.ApproverRef != nil {
			var approverId string
			user, err := resolvers.ResolveUser(ctx, e.kube, approver.ApproverRef)
			if err != nil {
				group, err := resolvers.ResolveGroup(ctx, e.kube, approver.ApproverRef)
				if err != nil {
					return err
				}
				if group.Status.Descriptor == nil {
					return errors.New("group descriptor is nil")
				}
				groupResource, err := groups.Get(ctx, e.azCli, groups.GetOptions{
					Organization:    project.Spec.Organization,
					GroupDescriptor: helpers.String(group.Status.Descriptor),
				})
				if err != nil {
					return err
				}
				approverId = groupResource.OriginID
			} else {
				if user.Status.Descriptor == nil {
					return errors.New("user descriptor is nil")
				}
				userResource, err := users.Get(ctx, e.azCli, users.GetOptions{
					Organization:   project.Spec.Organization,
					UserDescriptor: helpers.String(user.Status.Descriptor),
				})
				if err != nil {
					return err
				}
				approverId = helpers.String(userResource.OriginID)
			}
			approvers = append(approvers, checkconfiguration.Approver{
				ID:          approverId,
				DisplayName: approver.ApproverRef.Name,
			})
		}
	}

	var res *checkconfiguration.CheckConfiguration
	switch cr.Spec.Type {
	case string(CheckConfigurationTypeApproval):
		res, err = checkconfiguration.Create[checkconfiguration.Approval](ctx, e.azCli, checkconfiguration.CreateOptions[checkconfiguration.Approval]{
			Organization: project.Spec.Organization,
			Project:      project.Status.Id,
			CheckRes: checkconfiguration.Approval{
				Settings: checkconfiguration.ApprovalSettings{
					Approvers:                 approvers,
					Instructions:              cr.Spec.ApprovalSettings.Instructions,
					MinRequiredApprovers:      cr.Spec.ApprovalSettings.MinRequiredApprovers,
					ExecutionOrder:            cr.Spec.ApprovalSettings.ExecutionOrder,
					BlockedApprovers:          cr.Spec.ApprovalSettings.BlockedApprovers,
					RequesterCannotBeApprover: cr.Spec.ApprovalSettings.RequesterCannotBeApprover,
				},
				Timeout: cr.Spec.Timeout,
				Type: checkconfiguration.Type{
					ID:   CheckConfigurationTypeApproval.GetIdFromType(),
					Name: CheckConfigurationTypeApproval.String(),
				},
				Resource: checkconfiguration.Resource{
					ID:   helpers.String(resourceId),
					Type: strings.ToLower(cr.Spec.Resource.Type), // type of the resource MUST be lowercase!
					Name: cr.Spec.Resource.ResourceRef.Name,
				},
			},
		})
		if err != nil {
			return err
		}
	case string(CheckConfigurationTypeTaskCheck):
		var jsonMap map[string]interface{}
		json.Unmarshal([]byte(cr.Spec.TaskCheckSettings.Inputs), &jsonMap)
		//inputs, err := json.Marshal(jsonMap)
		if err != nil {
			return err
		}
		res, err = checkconfiguration.Create[checkconfiguration.TaskCheck](ctx, e.azCli, checkconfiguration.CreateOptions[checkconfiguration.TaskCheck]{
			Organization: project.Spec.Organization,
			Project:      project.Status.Id,
			CheckRes: checkconfiguration.TaskCheck{
				Resource: checkconfiguration.Resource{
					ID:   helpers.String(resourceId),
					Type: strings.ToLower(cr.Spec.Resource.Type), // type of the resource MUST be lowercase!
					Name: cr.Spec.Resource.ResourceRef.Name,
				},
				Timeout: cr.Spec.Timeout,
				Type: checkconfiguration.Type{
					ID:   CheckConfigurationTypeTaskCheck.GetIdFromType(),
					Name: CheckConfigurationTypeTaskCheck.String(),
				},
				Settings: checkconfiguration.TaskCheckSettings{
					Inputs:              jsonMap,
					LinkedVariableGroup: cr.Spec.TaskCheckSettings.LinkedVariableGroup,
					RetryInterval:       cr.Spec.TaskCheckSettings.RetryInterval,
					DisplayName:         cr.Spec.TaskCheckSettings.DisplayName,
					DefinitionRef: checkconfiguration.DefinitionRef{
						Id:      cr.Spec.TaskCheckSettings.DefinitionRef.Id,
						Name:    cr.Spec.TaskCheckSettings.DefinitionRef.Name,
						Version: cr.Spec.TaskCheckSettings.DefinitionRef.Version,
					},
				},
			},
		})
	case string(CheckConfigurationTypeExtendsCheck):
		var extendedSettings []checkconfiguration.ExtendsCheckSetting
		for _, extendsCheck := range cr.Spec.ExtendsCheckSettings {
			extendedSettings = append(extendedSettings, checkconfiguration.ExtendsCheckSetting{
				RepositoryType: extendsCheck.RepositoryType,
				RepositoryName: extendsCheck.RepositoryName,
				RepositoryRef:  extendsCheck.RepositoryRef,
				TemplatePath:   extendsCheck.TemplatePath,
			})
		}
		res, err = checkconfiguration.Create[checkconfiguration.ExtendsCheck](ctx, e.azCli, checkconfiguration.CreateOptions[checkconfiguration.ExtendsCheck]{
			Organization: project.Spec.Organization,
			Project:      project.Status.Id,
			CheckRes: checkconfiguration.ExtendsCheck{
				Type: checkconfiguration.Type{
					ID:   CheckConfigurationTypeExtendsCheck.GetIdFromType(),
					Name: CheckConfigurationTypeExtendsCheck.String(),
				},
				Resource: checkconfiguration.Resource{
					Type: strings.ToLower(cr.Spec.Resource.Type), // type of the resource MUST be lowercase!
					ID:   helpers.String(resourceId),
					Name: cr.Spec.Resource.ResourceRef.Name,
				},
				Settings: checkconfiguration.ExtendsCheckSettings{
					ExtendsChecks: extendedSettings,
				},
			},
		})
	}
	if err != nil {
		return err
	}

	cr.Status.ID = helpers.StringPtr(fmt.Sprintf("%v", res.ID))
	e.kube.Status().Update(ctx, cr)

	e.log.Debug("Creating CheckConfiguration", "organization", project.Spec.Organization, "project", project.Status.Id, "checkId", res.ID)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "CheckConfigurationCreating", "CheckConfiguration creating")

	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // Cannot be implemented because CheckConfiguration "settings" are not returned by the Azure DevOps API
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*checkconfigurations1alpha1.CheckConfiguration)
	if !ok {
		return errors.New(errNotCR)
	}

	e.log.Info("Deleting resource")

	project, err := resolvers.ResolveTeamProject(ctx, e.kube, cr.Spec.ProjectRef)
	if err != nil {
		return err
	}

	err = checkconfiguration.Delete(ctx, e.azCli, checkconfiguration.DeleteOptions{
		Organization: project.Spec.Organization,
		Project:      project.Status.Id,
		CheckId:      helpers.String(cr.Status.ID),
	})
	return err
}

// RemoveFormatting removes formatting (newlines and indentation) from a string
func RemoveFormatting(input string) string {
	// Replace newlines with spaces
	re := regexp.MustCompile(`\r?\n`)
	noNewlines := re.ReplaceAllString(input, " ")

	// Remove extra spaces
	re = regexp.MustCompile(`\s+`)
	noExtraSpaces := re.ReplaceAllString(noNewlines, " ")

	return strings.TrimSpace(noExtraSpaces)
}
