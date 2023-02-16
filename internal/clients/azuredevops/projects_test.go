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

func TestCreateProject(t *testing.T) {
	cli := setupClient()

	res, err := CreateProject(context.TODO(), cli, CreateProjectOpts{
		Organization: os.Getenv("ORG"),
		TeamProject: &TeamProject{
			Name:        helpers.StringPtr("ProjectCreatedViaRestApi"),
			Description: helpers.StringPtr("Test Project Created by Go!"),
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
		ProjectId:    "e25a896e-a4a8-4046-b703-96ccd4d94deb",
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
	env, _ := dotenv.FromFile("../../../../../.env")
	dotenv.PutInEnv(env, false)

	httpClient := httplib.CreateHTTPClient(httplib.CreateHTTPClientOpts{})

	return NewClient(httpClient, Options{
		Verbose: true,
		BaseURL: os.Getenv("BASE_URL"),
		Token:   os.Getenv("TOKEN"),
	})
}
