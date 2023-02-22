//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"fmt"
	"os"
	"testing"

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
