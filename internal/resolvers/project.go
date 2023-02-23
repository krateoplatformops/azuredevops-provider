package resolvers

import (
	"context"
	"fmt"

	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func ResolveTeamProject(ctx context.Context, kube client.Client, ref *rtv1.Reference) (projects.TeamProject, error) {
	res := projects.TeamProject{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &res)
	return res, err
}
