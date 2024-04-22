package resolvers

import (
	"context"
	"fmt"
	"strings"

	endpoint "github.com/krateoplatformops/azuredevops-provider/apis/endpoints/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ResolveEndpoint(ctx context.Context, kube client.Client, ref *rtv1.Reference) (*endpoint.Endpoint, error) {
	res := &endpoint.Endpoint{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, res)
	return res, err
}

func FindEndpointRef(ctx context.Context, kube client.Client, id string) (*rtv1.Reference, error) {
	list := &endpoint.EndpointList{}
	err := kube.List(ctx, list)
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no Endpoint referenced")
	}

	for _, v := range list.Items {
		sid := fmt.Sprintf("%v", helpers.String(v.Status.Id))
		if strings.EqualFold(sid, id) {
			return &rtv1.Reference{
				Name:      v.ObjectMeta.GetName(),
				Namespace: v.GetObjectMeta().GetNamespace(),
			}, nil
		}
	}
	return nil, fmt.Errorf("no Endpoint referenced with id: %s", id)
}
