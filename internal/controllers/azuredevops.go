package controllers

import (
	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/pipeline"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/pipelinepermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/project"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/repository"
	"github.com/krateoplatformops/azuredevops-provider/internal/controllers/run"
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
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
