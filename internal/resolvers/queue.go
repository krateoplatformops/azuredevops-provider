package resolvers

import (
	"context"
	"fmt"

	queue "github.com/krateoplatformops/azuredevops-provider/apis/queues/v1alpha1"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ResolveQueue(ctx context.Context, kube client.Client, ref *rtv1.Reference) (*queue.Queue, error) {
	res := &queue.Queue{}
	if ref == nil {
		return res, fmt.Errorf("no %s referenced", res.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, res)
	return res, err
}

func FindQueueRef(ctx context.Context, kube client.Client, id string) (*rtv1.Reference, error) {
	list := &queue.QueueList{}
	err := kube.List(ctx, list)
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no Queue referenced")
	}

	for _, v := range list.Items {
		sid := fmt.Sprintf("%v", v.Status.Id)
		if sid == id {
			return &rtv1.Reference{
				Name:      v.ObjectMeta.GetName(),
				Namespace: v.GetObjectMeta().GetNamespace(),
			}, nil
		}
	}
	return nil, err
}
