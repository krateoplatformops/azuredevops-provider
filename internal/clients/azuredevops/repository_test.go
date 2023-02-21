//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"fmt"
	"os"
	"testing"

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

	fmt.Printf("%v\n", res)
}

func TestGetRepository(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.GetRepository(context.TODO(), GetRepositoryOptions{
		Organization: os.Getenv("ORG"),
		ProjectId:    os.Getenv("PROJECT_ID"),
		RepositoryId: os.Getenv("REPO_NAME"),
	})
	if err != nil {
		if !httplib.IsNotFoundError(err) {
			t.Fatal(err)
		}
	}

	t.Logf("%s (id: %s)\n", *res.Name, *res.Id)
}
