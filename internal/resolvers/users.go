package resolvers

import (
	"context"
	"fmt"

	users "github.com/krateoplatformops/azuredevops-provider/apis/users/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func ResolveUser(ctx context.Context, kube client.Client, ref *rtv1.Reference) (*users.Users, error) {
	res := &users.Users{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, res)
	return res, err
}
