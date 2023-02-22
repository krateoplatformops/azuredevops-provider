package teamproject

import (
	"context"
	"fmt"

	connectorconfigv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/connectorconfig/v1alpha1"
	teamprojectv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/teamproject/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/meta"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// annotationKeyOperation is the key in the annotations map of a
	// async operation for the name of the resource to be created.
	annotationKeyOperation = "krateo.io/opid"
)

func (c *connector) clientOptions(ctx context.Context, ref *teamprojectv1alpha1.ConnectorConfigSelector) (azuredevops.ClientOptions, error) {
	opts := azuredevops.ClientOptions{}

	if ref == nil {
		return opts, errors.New("no ConnectorConfig referenced")
	}

	cfg := connectorconfigv1alpha1.ConnectorConfig{}
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

func teamProjectFromSpec(spec *teamprojectv1alpha1.TeamProjectSpec) *azuredevops.TeamProject {
	visibility := azuredevops.VisibilityPrivate
	if spec.Visibility != nil {
		visibility = azuredevops.ProjectVisibility(helpers.String(spec.Visibility))
	}

	res := &azuredevops.TeamProject{
		Name:        spec.Name,
		Description: helpers.StringPtr(spec.Description),
		Visibility:  visibility,
	}

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

// getOperationAnnotation returns the azuredevops operation annotation.
func getOperationAnnotation(o metav1.Object) string {
	return o.GetAnnotations()[annotationKeyOperation]
}

// setOperationAnnotation sets the azuredevops operation annotation.
func setOperationAnnotation(o metav1.Object, identifier string) {
	meta.AddAnnotations(o, map[string]string{annotationKeyOperation: identifier})
}

// deleteOperationAnnotation delete the azuredevops operation annotation.
func deleteOperationAnnotation(o metav1.Object) {
	meta.RemoveAnnotations(o, annotationKeyOperation)
}
