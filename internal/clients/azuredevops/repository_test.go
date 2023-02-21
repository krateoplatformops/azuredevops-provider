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
