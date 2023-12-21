package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	checkconfigurations "github.com/krateoplatformops/azuredevops-provider/apis/checkconfigurations/v1alpha1"
	connectorconfigs "github.com/krateoplatformops/azuredevops-provider/apis/connectorconfigs/v1alpha1"
	endpoints "github.com/krateoplatformops/azuredevops-provider/apis/endpoints/v1alpha1"
	environments "github.com/krateoplatformops/azuredevops-provider/apis/environments/v1alpha1"
	feedpermissions "github.com/krateoplatformops/azuredevops-provider/apis/feedpermissions/v1alpha1"
	feeds "github.com/krateoplatformops/azuredevops-provider/apis/feeds/v1alpha1"
	groups "github.com/krateoplatformops/azuredevops-provider/apis/groups/v1alpha1"
	pipelinepermissionsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha1"
	pipelinepermissionsv1alpha2 "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha2"
	pipelines "github.com/krateoplatformops/azuredevops-provider/apis/pipelines/v1alpha1"
	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	queues "github.com/krateoplatformops/azuredevops-provider/apis/queues/v1alpha1"
	repositories "github.com/krateoplatformops/azuredevops-provider/apis/repositories/v1alpha1"
	repositorypermissions "github.com/krateoplatformops/azuredevops-provider/apis/repositorypermissions/v1alpha1"
	runs "github.com/krateoplatformops/azuredevops-provider/apis/runs/v1alpha1"
	teams "github.com/krateoplatformops/azuredevops-provider/apis/teams/v1alpha1"
	users "github.com/krateoplatformops/azuredevops-provider/apis/users/v1alpha1"
)

func init() {
	AddToSchemes = append(AddToSchemes,
		connectorconfigs.SchemeBuilder.AddToScheme,
		projects.SchemeBuilder.AddToScheme,
		repositories.SchemeBuilder.AddToScheme,
		pipelines.SchemeBuilder.AddToScheme,
		runs.SchemeBuilder.AddToScheme,
		pipelinepermissionsv1alpha1.SchemeBuilder.AddToScheme,
		pipelinepermissionsv1alpha2.SchemeBuilder.AddToScheme,
		teams.SchemeBuilder.AddToScheme,
		feeds.SchemeBuilder.AddToScheme,
		queues.SchemeBuilder.AddToScheme,
		endpoints.SchemeBuilder.AddToScheme,
		feedpermissions.SchemeBuilder.AddToScheme,
		environments.SchemeBuilder.AddToScheme,
		repositories.SchemeBuilder.AddToScheme,
		repositorypermissions.SchemeBuilder.AddToScheme,
		groups.SchemeBuilder.AddToScheme,
		users.SchemeBuilder.AddToScheme,
		checkconfigurations.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
