package resolvers

import (
	"context"
	"fmt"

	projects "github.com/krateoplatformops/azuredevops-provider/apis/projects/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func ResolveTeamProject(ctx context.Context, kube client.Client, ref *rtv1.Reference) (*projects.TeamProject, error) {
	res := &projects.TeamProject{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, res)
	return res, err
}

func FindTeamProject(ctx context.Context, kube client.Client, projectId string) (*projects.TeamProject, error) {
	res := &projects.TeamProject{}
	list := &projects.TeamProjectList{}
	err := kube.List(ctx, list)

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no teamproject referenced")
	}

	for _, v := range list.Items {
		if v.Status.Id == projectId {
			return &v, nil
		}
	}
	return res, err
}
func FindTeamProjectRef(ctx context.Context, kube client.Client, projectId string) (*rtv1.Reference, error) {
	list := &projects.TeamProjectList{}
	err := kube.List(ctx, list)
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no teamproject referenced")
	}

	for _, v := range list.Items {
		if v.Status.Id == projectId {
			return &rtv1.Reference{
				Name:      v.ObjectMeta.GetName(),
				Namespace: v.GetObjectMeta().GetNamespace(),
			}, nil
		}
	}
	return nil, err
}
