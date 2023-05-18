package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	connectorconfigs "github.com/krateoplatformops/azuredevops-provider/apis/connectorconfigs/v1alpha1"
	pipelinepermissions "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha1"
	pipelines "github.com/krateoplatformops/azuredevops-provider/apis/pipelines/v1alpha1"

	endpoints "github.com/krateoplatformops/azuredevops-provider/apis/endpoints/v1alpha1"
	feeds "github.com/krateoplatformops/azuredevops-provider/apis/feeds/v1alpha1"
	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	queues "github.com/krateoplatformops/azuredevops-provider/apis/queues/v1alpha1"
	repositories "github.com/krateoplatformops/azuredevops-provider/apis/repositories/v1alpha1"
	runs "github.com/krateoplatformops/azuredevops-provider/apis/runs/v1alpha1"
)

func init() {
	AddToSchemes = append(AddToSchemes,
		connectorconfigs.SchemeBuilder.AddToScheme,
		projects.SchemeBuilder.AddToScheme,
		repositories.SchemeBuilder.AddToScheme,
		pipelines.SchemeBuilder.AddToScheme,
		runs.SchemeBuilder.AddToScheme,
		pipelinepermissions.SchemeBuilder.AddToScheme,
		feeds.SchemeBuilder.AddToScheme,
		queues.SchemeBuilder.AddToScheme,
		endpoints.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
