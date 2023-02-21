//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/dotenv"
	"github.com/lucasepe/httplib"
)

func TestListProjects(t *testing.T) {
	cli := createAzureDevopsClient()

	var continutationToken string
	for {
		top := int(4)
		res, err := cli.ListProjects(context.TODO(), ListProjectsOpts{
			Organization:      os.Getenv("ORG"),
			StateFilter:       (*ProjectState)(helpers.StringPtr("all")),
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
			t.Logf("%s (id: %s)", el.Name, *el.Id)
		}

		continutationToken = *res.ContinuationToken
		if continutationToken == "" {
			break
		}
	}
}

func TestCreateProject(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.CreateProject(context.TODO(), CreateProjectOpts{
		Organization: os.Getenv("ORG"),
		TeamProject: &TeamProject{
			Name: os.Getenv("PROJECT_NAME"),
			//Description: helpers.StringPtr("Sorry for the Spam but I need to let the continuation token appear..."),
			Capabilities: &Capabilities{
				&Versioncontrol{
					SourceControlType: "Git",
				},
				&ProcessTemplate{
					TemplateTypeId: os.Getenv("TEMPLATE_ID"),
				},
			},
		},
	})
	if err != nil {
		var apierr *APIError
		if errors.As(err, &apierr) {
			fmt.Println(apierr.Error())
		}
	}

	fmt.Printf("%v\n", res)
}

func TestGetProject(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.GetProject(context.TODO(), GetProjectOpts{
		Organization: os.Getenv("ORG"),
		ProjectId:    os.Getenv("PROJECT_ID"),
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%v\n", res)
}

func TestDeleteProject(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.DeleteProject(context.TODO(), DeleteProjectOpts{
		Organization: os.Getenv("ORG"),
		ProjectId:    os.Getenv("PROJECT_ID"),
	})
	if err != nil {
		if !httplib.HasStatusErr(err, http.StatusNotFound) {
			t.Fatal(err)
		}
	}

	if res != nil {
		t.Logf("operationId: %s", res.Id)
	}
}

func TestFindProject(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := cli.FindProject(context.TODO(), FindProjectsOpts{
		Organization: os.Getenv("ORG"),
		Name:         os.Getenv("PROJECT_NAME"),
	})
	if err != nil {
		if !httplib.HasStatusErr(err, http.StatusNotFound) {
			t.Fatal(err)
		}
	}

	if res != nil {
		fmt.Printf("%+v\n", helpers.String(res.Id))
	}
}

func createAzureDevopsClient() *Client {
	env, _ := dotenv.FromFile("../../../.env")
	dotenv.PutInEnv(env, false)

	return NewClient(ClientOptions{
		Verbose: false,
		BaseURL: os.Getenv("BASE_URL"),
		Token:   os.Getenv("TOKEN"),
	})
}
