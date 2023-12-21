package resolvers

import (
	"context"
	"fmt"
	"strings"

	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceType string

const (
	GitRepository ResourceType = "repository"
	Environment   ResourceType = "environment"
	Queue         ResourceType = "queue"
	TeamProject   ResourceType = "teamproject"
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

func ResolveResourceId(ctx context.Context, cli client.Client, ref *rtv1.Reference, ty string) (*string, error) {
	if ref == nil {
		return nil, fmt.Errorf("no resource referenced")
	}
	switch strings.ToLower(ty) {
	case string(GitRepository):
		repo, err := ResolveGitRepository(ctx, cli, ref)
		if err != nil {
			return nil, err
		}
		proj, err := ResolveTeamProject(ctx, cli, repo.Spec.ProjectRef)
		ret := fmt.Sprintf("%s.%s", proj.Status.Id, repo.Status.Id)
		return helpers.StringPtr(ret), err
	case string(Environment):
		env, err := ResolveEnvironment(ctx, cli, ref)
		ret := fmt.Sprintf("%v", helpers.Int(env.Status.Id))
		return helpers.StringPtr(ret), err
	case string(Queue):
		que, err := ResolveQueue(ctx, cli, ref)
		ret := fmt.Sprintf("%v", helpers.Int(que.Status.Id))
		return helpers.StringPtr(ret), err
	}

	return nil, fmt.Errorf("no resource referenced of type %s", ty)
}
