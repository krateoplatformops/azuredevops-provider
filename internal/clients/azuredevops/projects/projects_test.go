//go:build integration
// +build integration

package projects

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"testing"

// 	"github.com/davecgh/go-spew/spew"
// 	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
// 	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
// 	"github.com/lucasepe/dotenv"
// 	"github.com/lucasepe/httplib"
// )

// func TestListProjects(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	var continutationToken string
// 	for {
// 		top := int(4)
// 		res, err := List(context.TODO(), cli, ListOptions{
// 			Organization:      os.Getenv("ORG"),
// 			StateFilter:       (*ProjectState)(helpers.StringPtr("all")),
// 			Top:               &top,
// 			ContinuationToken: &continutationToken,
// 		})
// 		if err != nil {
// 			var apierr *azuredevops.APIError
// 			if errors.As(err, &apierr) {
// 				fmt.Println(apierr.Error())
// 			}
// 			break
// 		}

// 		for _, el := range res.Value {
// 			t.Logf("%s (id: %s)", el.Name, *el.Id)
// 		}

// 		continutationToken = *res.ContinuationToken
// 		if continutationToken == "" {
// 			break
// 		}
// 	}
// }

// func TestCreateProject(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Create(context.TODO(), cli, CreateOptions{
// 		Organization: os.Getenv("ORG"),
// 		TeamProject: &TeamProject{
// 			Name: os.Getenv("PROJECT_NAME"),
// 			//Description: helpers.StringPtr("Sorry for the Spam but I need to let the continuation token appear..."),
// 			Capabilities: &Capabilities{
// 				&Versioncontrol{
// 					SourceControlType: "Git",
// 				},
// 				&ProcessTemplate{
// 					TemplateTypeId: os.Getenv("TEMPLATE_ID"),
// 				},
// 			},
// 		},
// 	})
// 	if err != nil {
// 		var apierr *azuredevops.APIError
// 		if errors.As(err, &apierr) {
// 			fmt.Println(apierr.Error())
// 		}
// 	}

// 	fmt.Printf("%v\n", res)
// }

// func TestGetProject(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Get(context.TODO(), cli, GetOptions{
// 		Organization: os.Getenv("ORG"),
// 		ProjectId:    os.Getenv("PROJECT_NAME"),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	spew.Dump(res)
// }

// func TestDeleteProject(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Delete(context.TODO(), cli, DeleteOptions{
// 		Organization: os.Getenv("ORG"),
// 		ProjectId:    "de69b1ba-ce86-4275-8d2c-653e4b354a7b", //os.Getenv("PROJECT_ID"),
// 	})
// 	if err != nil {
// 		if !httplib.HasStatusErr(err, http.StatusNotFound) {
// 			t.Fatal(err)
// 		}
// 	}

// 	if res != nil {
// 		t.Logf("operationId: %s", res.Id)
// 	}
// }

// func TestFindProject(t *testing.T) {
// 	cli := createAzureDevopsClient()

// 	res, err := Find(context.TODO(), cli, FindOptions{
// 		Organization: os.Getenv("ORG"),
// 		Name:         os.Getenv("PROJECT_NAME"),
// 	})
// 	if err != nil {
// 		if !httplib.HasStatusErr(err, http.StatusNotFound) {
// 			t.Fatal(err)
// 		}
// 	}

// 	if res != nil {
// 		fmt.Printf("%+v\n", helpers.String(res.Id))
// 	}
// }

// func createAzureDevopsClient() *azuredevops.Client {
// 	env, _ := dotenv.FromFile("../../../../.env")
// 	dotenv.PutInEnv(env, false)

// 	return azuredevops.NewClient(azuredevops.ClientOptions{
// 		Verbose: false,
// 		Token:   os.Getenv("TOKEN"),
// 	})
// }
