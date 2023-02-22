package v1alpha1

import (
	"context"

	"github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/reference"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this Pipeline.
func (mg *Pipeline) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.Project
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: helpers.String(mg.Spec.Project),
		Reference:    mg.Spec.PojectRef,
		To:           &v1alpha1.TeamProject{},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.Project")
	}
	mg.Spec.Project = helpers.StringPtr(rsp.ResolvedValue)
	mg.Spec.PojectRef = rsp.ResolvedReference

	return nil
}
