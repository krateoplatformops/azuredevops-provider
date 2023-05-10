//go:build integration
// +build integration

package pipelines

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/dotenv"
	"github.com/lucasepe/httplib"
)

func TestGetPipeline(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Get(context.TODO(), cli, GetOptions{
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

	res, err := Create(context.TODO(), cli, CreateOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		Pipeline: Pipeline{
			Name:   os.Getenv("PIPELINE_NAME"),
			Folder: os.Getenv("PIPELINE_FOLDER"),
			Configuration: &PipelineConfiguration{
				Type: ConfigurationYaml,
				Path: helpers.StringPtr("/azure-pipelines-helloworld.yaml"),
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
		res, err := List(context.TODO(), cli, ListOptions{
			Organization:      os.Getenv("ORG"),
			Project:           os.Getenv("PROJECT_NAME"),
			Top:               &top,
			ContinuationToken: &continutationToken,
		})
		if err != nil {
			var apierr *azuredevops.APIError
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

func TestFindPipeline(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Find(context.TODO(), cli, FindOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		Name:         os.Getenv("PIPELINE_NAME"),
	})
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(res)
}

func createAzureDevopsClient() *azuredevops.Client {
	env, _ := dotenv.FromFile("../../../../.env")
	dotenv.PutInEnv(env, false)

	return azuredevops.NewClient(azuredevops.ClientOptions{
		Verbose: false,
		BaseURL: os.Getenv("BASE_URL"),
		Token:   os.Getenv("TOKEN"),
	})
}
