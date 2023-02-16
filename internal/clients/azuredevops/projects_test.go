package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httplib"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/dotenv"
)

func TestListProjects(t *testing.T) {
	cli := setupClient()

	top := int(2)
	res, err := ListProjects(context.TODO(), cli, ListProjectsOpts{
		Organization: os.Getenv("ORG"),
		Top:          &top,
	})
	if err != nil {
		var apierr *APIError
		if errors.As(err, &apierr) {
			fmt.Println(apierr.Error())
		}
	}

	fmt.Printf("%v\n", res)
}

func TestCreateProject(t *testing.T) {
	cli := setupClient()

	res, err := CreateProject(context.TODO(), cli, CreateProjectOpts{
		Organization: os.Getenv("ORG"),
		TeamProject: &TeamProject{
			Name:        helpers.StringPtr("Project Created Via Rest Api"),
			Description: helpers.StringPtr("Test Project Created by Go! #2"),
			Capabilities: &map[string]map[string]string{
				"versioncontrol": {
					"sourceControlType": "Git",
				},
				"processTemplate": {
					"templateTypeId": "6b724908-ef14-45cf-84f8-768b5384da45",
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
	cli := setupClient()

	res, err := GetProject(context.TODO(), cli, GetProjectOpts{
		Organization: os.Getenv("ORG"),
		ProjectId:    "bdb1db89-f1ea-45a7-89c2-97eff028a5a8",
	})
	if err != nil {
		if IsNotFound(err) {
			fmt.Println("NOT FOUNDOK")
		}

		//var apierr *APIError
		//if errors.As(err, &apierr) {
		//	fmt.Println(apierr.Error())
		//}

	}

	fmt.Printf("%v\n", res)
}

func setupClient() *Client {
	env, _ := dotenv.FromFile("../../../.env")
	dotenv.PutInEnv(env, false)

	httpClient := httplib.CreateHTTPClient(httplib.CreateHTTPClientOpts{})

	return NewClient(httpClient, Options{
		Verbose: true,
		BaseURL: os.Getenv("BASE_URL"),
		Token:   os.Getenv("TOKEN"),
	})
}
