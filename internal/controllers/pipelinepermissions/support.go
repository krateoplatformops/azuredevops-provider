package pipelinepermissions

import (
	"context"
	"fmt"

	pipelineperm "github.com/krateoplatformops/azuredevops-provider/apis/pipelinepermissions/v1alpha2"
	pipelinespermissions "github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/pipelinespermissions"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func checkPipelinePermission(ctx context.Context, kube client.Client, currentPip []pipelineperm.PipelineAuthorization, observedPip []pipelinespermissions.PipelinePermission) (bool, error) {
	for _, current := range currentPip {
		currentPip, err := resolvers.ResolvePipeline(ctx, kube, current.PipelineRef)
		if err != nil {
			return false, fmt.Errorf("cannot resolve pipeline: %w", err)
		}
		found := false
		for _, observed := range observedPip {
			if current.Authorized && helpers.String(currentPip.Status.Id) == observed.GetId() {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}
