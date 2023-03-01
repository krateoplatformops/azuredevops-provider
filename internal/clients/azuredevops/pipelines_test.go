//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

func TestGetPipeline(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.GetPipeline(context.TODO(), GetPipelineOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_ID"),
		PipelineId:   os.Getenv("PIPELINE_ID"),
	})
	if err != nil {
		if httplib.IsNotFoundError(err) {
			return
		}
		t.Fatal(err)
	}

	t.Logf("%s (id: %d)\n", res.Name, *res.Id)
}

func TestCreatePipeline(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.CreatePipeline(context.TODO(), CreatePipelineOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		Pipeline: Pipeline{
			Name:   os.Getenv("PIPELINE_NAME"),
			Folder: os.Getenv("PIPELINE_FOLDER"),
			Configuration: &PipelineConfiguration{
				Type: ConfigurationYaml,
				Path: helpers.StringPtr("/azure-pipelines.yml"),
				Repository: &BuildRepository{
					Id:   os.Getenv("REPO_ID"),
					Name: os.Getenv("REPO_NAME"),
					Type: BuildRepositoryAzureReposGit,
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("id: %d\n", *res.Id)
}

func TestListPipelines(t *testing.T) {
	cli := createAzureDevopsClient()

	var continutationToken string
	for {
		top := int(4)
		res, err := cli.ListPipelines(context.TODO(), ListPipelinesOptions{
			Organization:      os.Getenv("ORG"),
			Project:           os.Getenv("PROJECT_NAME"),
			Top:               &top,
			ContinuationToken: &continutationToken,
		})
		if err != nil {
			var apierr *APIError
			if errors.As(err, &apierr) {
				fmt.Println(apierr.Error())
			}
			break
		}

		for _, el := range res.Value {
			t.Logf("%s (id: %d)", el.Name, *el.Id)
		}

		continutationToken = *res.ContinuationToken
		if continutationToken == "" {
			break
		}
	}
}
