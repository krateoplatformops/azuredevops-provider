//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/lucasepe/httplib"
)

func TestDeleteDefinition(t *testing.T) {
	cli := createAzureDevopsClient()

	err := cli.DeleteDefinition(context.TODO(), DeleteDefinitionOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		DefinitionId: "28",
	})
	if err != nil {
		if !httplib.HasStatusErr(err, http.StatusNotFound) {
			t.Fatal(err)
		}
	}
}