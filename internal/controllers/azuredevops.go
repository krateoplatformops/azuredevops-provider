package controllers

import (
	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/checkconfigurations"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/endpoints"
	environments "github.com/krateoplatformops/azuredevops-provider/internal/controllers/enviroments"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/feedpermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/feeds"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/groups"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/pipeline"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/pipelinepermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/project"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/queues"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/repository"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/repositorypermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/run"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/teams"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/users"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/variablegroups"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		project.Setup,
		repository.Setup,
		pipeline.Setup,
		run.Setup,
		pipelinepermissions.Setup,
		feeds.Setup,
		queues.Setup,
		endpoints.Setup,
		feedpermissions.Setup,
		environments.Setup,
		repositorypermissions.Setup,
		checkconfigurations.Setup,
		teams.Setup,
		groups.Setup,
		users.Setup,
		variablegroups.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
