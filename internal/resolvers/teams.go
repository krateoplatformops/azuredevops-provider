package resolvers

import (
	"context"
	"fmt"

	teams "github.com/krateoplatformops/azuredevops-provider/apis/teams/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func ResolveTeam(ctx context.Context, kube client.Client, ref *rtv1.Reference) (*teams.Team, error) {
	res := &teams.Team{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, res)
	return res, err
}

func ResolveTeamListDescriptors(ctx context.Context, kube client.Client, refs []rtv1.Reference) (list []string, err error) {
	var descriptor *string
	for _, ref := range refs {
		res, err := ResolveTeam(ctx, kube, &ref)
		descriptor = res.Status.Descriptor
		if err != nil {
			return nil, err
		}
		if descriptor != nil {
			list = append(list, helpers.String(descriptor))
		}
	}
	return list, nil
}
