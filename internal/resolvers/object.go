package resolvers

import (
	"context"

	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetFinderFromType(ty string) func(context.Context, client.Client, string) (*rtv1.Reference, error) {
	switch ty {
	case "teamproject":
		return FindTeamProjectRef
	case "repository":
		return FindRepositoryRef
	case "queue":
		return FindQueueRef
	case "environment":
		return FindEnvironmentRef
	}
	return nil
}
