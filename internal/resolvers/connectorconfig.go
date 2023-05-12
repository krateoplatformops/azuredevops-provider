package resolvers

import (
	"context"
	"fmt"

	connectorconfigs "github.com/krateoplatformops/azuredevops-provider/apis/connectorconfigs/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func ResolveConnectorConfig(ctx context.Context, kube client.Client, ref *rtv1.Reference) (azuredevops.ClientOptions, error) {
	opts := azuredevops.ClientOptions{}

	cfg := connectorconfigs.ConnectorConfig{}
	if ref == nil {
		return opts, fmt.Errorf("no %s referenced", cfg.Kind)
	}

	err := kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &cfg)
	if err != nil {
		return opts, errors.Wrapf(err, "cannot get %s connector config", ref.Name)
	}

	csr := cfg.Spec.Credentials.SecretRef
	if csr == nil {
		return opts, fmt.Errorf("no credentials secret referenced")
	}

	sec := corev1.Secret{}
	err = kube.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, &sec)
	if err != nil {
		return opts, errors.Wrapf(err, "cannot get %s secret", ref.Name)
	}

	token, err := resource.GetSecret(ctx, kube, csr.DeepCopy())
	if err != nil {
		return opts, err
	}

	opts.Token = token
	opts.Verbose = false

	return opts, nil
}
