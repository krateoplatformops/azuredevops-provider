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
	"github.com/lucasepe/dotenv"
)

func TestListProjects(t *testing.T) {
	cli := setupClient()

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

		continutationToken = *res.ContinuationToken
		fmt.Printf("TOKEN => %v\n", continutationToken)
		if continutationToken == "" {
			break
		}
	}
}

func TestCreateProject(t *testing.T) {
	cli := setupClient()

	for i := 0; i < 20; i++ {
		res, err := cli.CreateProject(context.TODO(), CreateProjectOpts{
			Organization: os.Getenv("ORG"),
			TeamProject: &TeamProject{
				Name:        helpers.StringPtr(fmt.Sprintf("Created by Go nr.%d", i)),
				Description: helpers.StringPtr("Sorry for the Spam but I need to let the continuation token appear..."),
				Capabilities: &Capabilities{
					&Versioncontrol{
						SourceControlType: helpers.StringPtr("Git"),
					},
					&ProcessTemplate{
						TemplateTypeId: helpers.StringPtr("6b724908-ef14-45cf-84f8-768b5384da45"),
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
}

func TestGetProject(t *testing.T) {
	cli := setupClient()

	res, err := cli.GetProject(context.TODO(), GetProjectOpts{
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

func TestDeleteProject(t *testing.T) {
	cli := setupClient()

	res, err := cli.DeleteProject(context.TODO(), DeleteProjectOpts{
		Organization: os.Getenv("ORG"),
		ProjectId:    "401a7ba2-3043-4163-89e9-1a7707a41610",
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

	return NewClient(ClientOptions{
		Verbose: false,
		BaseURL: os.Getenv("BASE_URL"),
		Token:   os.Getenv("TOKEN"),
	})
}
