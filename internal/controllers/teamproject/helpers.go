package teamproject

import (
	"context"
	"fmt"
	"strings"

	teamprojectv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/teamproject/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func teamProjectFromSpec(spec *teamprojectv1alpha1.TeamProjectSpec) *azuredevops.TeamProject {
	visibility := azuredevops.VisibilityPrivate
	if spec.Visibility != nil {
		visibility = azuredevops.ProjectVisibility(helpers.String(spec.Visibility))
	}

	res := &azuredevops.TeamProject{
		Name:        helpers.StringPtr(spec.Name),
		Description: helpers.StringPtr(spec.Description),
		Visibility:  &visibility,
	}

	if spec.Capabilities != nil {
		res.Capabilities = &azuredevops.Capabilities{}
		if spec.Capabilities.Versioncontrol != nil {
			res.Capabilities.Versioncontrol = &azuredevops.Versioncontrol{
				SourceControlType: spec.Capabilities.Versioncontrol.SourceControlType,
			}
		}

		if spec.Capabilities.ProcessTemplate != nil {
			res.Capabilities.ProcessTemplate = &azuredevops.ProcessTemplate{
				TemplateTypeId: spec.Capabilities.ProcessTemplate.TemplateTypeId,
			}
		}
	}

	return res
}

// conditionFromOperationReference returns a condition that indicates
// the TeamProject resource is not currently available for use.
func conditionFromOperationReference(opref *azuredevops.OperationReference) rtv1.Condition {
	if opref == nil {
		return rtv1.Unavailable()
	}

	res := rtv1.Condition{
		Type:               rtv1.TypeReady,
		LastTransitionTime: metav1.Now(),
		Reason:             rtv1.ConditionReason(opref.Status),
	}

	switch s := opref.Status; {
	case s == azuredevops.StatusSucceeded:
		res.Status = corev1.ConditionTrue
	default:
		res.Status = corev1.ConditionFalse
	}

	return res
}

func findTeamProject(ctx context.Context, cli *azuredevops.Client, org, name string) (azuredevops.TeamProject, error) {
	var continutationToken string
	for {
		top := int(30)
		filter := azuredevops.StateWellFormed
		res, err := cli.ListProjects(ctx, azuredevops.ListProjectsOpts{
			Organization:      org,
			StateFilter:       &filter,
			Top:               &top,
			ContinuationToken: &continutationToken,
		})
		if err != nil {
			fmt.Printf("err => %v\n\n", err)
			return azuredevops.TeamProject{}, err
		}

		for _, el := range res.Value {
			fmt.Printf("name: %s, id: %s\n\n", *el.Name, *el.Id)
			if strings.EqualFold(*el.Name, name) {
				return el, nil
			}
		}

		continutationToken = *res.ContinuationToken
		if continutationToken == "" {
			break
		}
	}

	return azuredevops.TeamProject{}, fmt.Errorf("project '%s' not found", name)
}
