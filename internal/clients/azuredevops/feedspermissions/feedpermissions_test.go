//go:build integration
// +build integration

package feedpermissions

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/davecgh/go-spew/spew"
// 	"github.com/lucasepe/httplib"
// )

// func TestGet(t *testing.T) {
// 	cli := createAzureDevopsClient()
// 	//cli.SetVerbose(true)
// 	res, err := Get(context.TODO(), cli, GetOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_ID"),
// 		ResourceType: "repository",
// 		ResourceId:   fmt.Sprintf("%s.%s", os.Getenv("PROJECT_ID"), os.Getenv("REPO_ID")),
// 	})
// 	if err != nil {
// 		if httplib.IsNotFoundError(err) {
// 			return
// 		}
// 		t.Fatal(err)
// 	}

// 	spew.Dump(res)
// }

// func TestUpdate(t *testing.T) {
// 	cli := createAzureDevopsClient()
// 	cli.SetVerbose(true)
// 	res, err := Update(context.TODO(), cli, UpdateOptions{
// 		Organization: os.Getenv("ORG"),
// 		Project:      os.Getenv("PROJECT_ID"),
// 		ResourceType: "repository",
// 		ResourceId:   fmt.Sprintf("%s.%s", os.Getenv("PROJECT_ID"), os.Getenv("REPO_ID")),
// 		ResourceAuthorization: &ResourcePipelinePermissions{
// 			AllPipelines: &Permission{
// 				Authorized: true,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		if httplib.IsNotFoundError(err) {
// 			return
// 		}
// 		t.Fatal(err)
// 	}

// 	spew.Dump(res)
// }

// func createAzureDevopsClient() *azuredevops.Client {
// 	env, _ := dotenv.FromFile("../../../../.env")
// 	dotenv.PutInEnv(env, false)

// 	return azuredevops.NewClient(azuredevops.ClientOptions{
// 		Verbose: false,
// 		Token:   os.Getenv("TOKEN"),
// 	})
// }
