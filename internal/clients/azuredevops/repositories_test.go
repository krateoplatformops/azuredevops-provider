//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

func TestCreateRepository(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.CreateRepository(context.TODO(), CreateRepositoryOptions{
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

	_, err = cli.CreatePush(context.TODO(), GitPushOptions{
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

	res, err := cli.GetRepository(context.TODO(), GetRepositoryOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_ID"),
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

	repo, err := cli.GetRepository(context.TODO(), GetRepositoryOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_ID"),
		Repository:   os.Getenv("REPO_NAME"),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	err = cli.DeleteRepository(context.TODO(), DeleteRepositoryOptions{
		Organization: os.Getenv("ORG"),
		Project:      helpers.String(repo.Project.Id),
		RepositoryId: helpers.String(repo.Id),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	err = cli.DeleteRepositoryFromRecycleBin(context.TODO(), DeleteRepositoryOptions{
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

	res, err := cli.FindRepository(context.TODO(), FindRepositoryOptions{
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
