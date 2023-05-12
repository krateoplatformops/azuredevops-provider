//go:build integration
// +build integration

package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/dotenv"
	"github.com/lucasepe/httplib"
)

func TestListRepositorires(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := List(context.TODO(), cli, ListOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, el := range res.Value {
		t.Logf("%s (id: %s)", *el.Name, *el.Id)
	}
}

func TestCreateRepository(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Create(context.TODO(), cli, CreateOptions{
		Organization: os.Getenv("ORG"),
		ProjectId:    os.Getenv("PROJECT_ID"),
		Name:         os.Getenv("REPO_NAME"),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	fmt.Printf("%s\n", helpers.String(res.Id))

	defaultBranch := helpers.String(res.DefaultBranch)
	if len(defaultBranch) == 0 {
		defaultBranch = "refs/heads/master"
	}

	_, err = CreatePush(context.TODO(), cli, GitPushOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_ID"),
		RepositoryId: helpers.String(res.Id),
		Push: &GitPush{
			RefUpdates: &[]GitRefUpdate{
				{
					Name:        helpers.StringPtr(defaultBranch),
					OldObjectId: helpers.StringPtr("0000000000000000000000000000000000000000"),
				},
			},
			Commits: &[]GitCommitRef{
				{
					Comment: helpers.StringPtr("Initial commit."),
					Changes: []GitChange{
						{
							ChangeType: ChangeTypeAdd,
							Item: map[string]string{
								"path": "/README.md",
							},
							NewContent: &ItemContent{
								Content:     fmt.Sprintf("# %s", helpers.String(res.Name)),
								ContentType: ContentTypeRawText,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRepository(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Get(context.TODO(), cli, GetOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		Repository:   os.Getenv("REPO_NAME"),
	})
	if err != nil {
		if httplib.IsNotFoundError(err) {
			return
		}
		t.Fatal(err)
	}

	t.Logf("%s (id: %s)\n", *res.Name, *res.Id)

	t.Logf("%+v\n", *res.RemoteUrl)
}

func TestDeleteRepository(t *testing.T) {
	cli := createAzureDevopsClient()

	repo, err := Get(context.TODO(), cli, GetOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_ID"),
		Repository:   os.Getenv("REPO_NAME"),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	err = Delete(context.TODO(), cli, DeleteOptions{
		Organization: os.Getenv("ORG"),
		Project:      helpers.String(repo.Project.Id),
		RepositoryId: helpers.String(repo.Id),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	err = DeleteFromRecycleBin(context.TODO(), cli, DeleteOptions{
		Organization: os.Getenv("ORG"),
		Project:      helpers.String(repo.Project.Id),
		RepositoryId: helpers.String(repo.Id),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}
}

func TestFindRepository(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Find(context.TODO(), cli, FindOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_ID"),
		Name:         os.Getenv("REPO_NAME"),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	spew.Dump(res)
}

func createAzureDevopsClient() *azuredevops.Client {
	env, _ := dotenv.FromFile("../../../../.env")
	dotenv.PutInEnv(env, false)

	return azuredevops.NewClient(azuredevops.ClientOptions{
		Verbose: false,
		Token:   os.Getenv("TOKEN"),
	})
}
