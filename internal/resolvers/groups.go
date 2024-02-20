package resolvers

import (
	"context"
	"fmt"

	groups "github.com/krateoplatformops/azuredevops-provider/apis/groups/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
)

func ResolveGroup(ctx context.Context, kube client.Client, ref *rtv1.Reference) (*groups.Groups, error) {
	res := &groups.Groups{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, res)
	return res, err
}

func ResolveGroupListDescriptors(ctx context.Context, kube client.Client, refs []rtv1.Reference) (list []string, err error) {
	var descriptor *string
	for _, ref := range refs {
		res, err := ResolveGroup(ctx, kube, &ref)
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

func ResolveGroupAndTeamDescriptors(ctx context.Context, kube client.Client, groupRefs []rtv1.Reference, teamRefs []rtv1.Reference) (list []string, err error) {
	teamDescritors, err := ResolveTeamListDescriptors(ctx, kube, teamRefs)
	if err != nil {
		return nil, err
	}
	groupDescriptors, err := ResolveGroupListDescriptors(ctx, kube, groupRefs)
	if err != nil {
		return nil, err
	}
	var groupAndTeamDescriptors []string
	groupAndTeamDescriptors = append(groupAndTeamDescriptors, teamDescritors...)
	groupAndTeamDescriptors = append(groupAndTeamDescriptors, groupDescriptors...)
	return groupAndTeamDescriptors, nil
}
