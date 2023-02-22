package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	connectorconfigs "github.com/krateoplatformops/azuredevops-provider/apis/connectorconfigs/v1alpha1"
	pipelines "github.com/krateoplatformops/azuredevops-provider/apis/pipelines/v1alpha1"
	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	repositories "github.com/krateoplatformops/azuredevops-provider/apis/repositories/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		connectorconfigs.SchemeBuilder.AddToScheme,
		projects.SchemeBuilder.AddToScheme,
		repositories.SchemeBuilder.AddToScheme,
		pipelines.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
