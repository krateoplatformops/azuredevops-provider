//go:build integration
// +build integration

package pools

import (
	"context"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/dotenv"
)

func TestFind(t *testing.T) {
	cli := createAzureDevopsClient()

	res, err := Find(context.TODO(), cli, FindOptions{
		Organization: os.Getenv("ORG"),
		PoolName:     "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	if res == nil {
		return
	}

	spew.Dump(res)

}

func createAzureDevopsClient() *azuredevops.Client {
	env, _ := dotenv.FromFile("../../../../.env")
	dotenv.PutInEnv(env, false)

	return azuredevops.NewClient(azuredevops.ClientOptions{
		Verbose: false,
		Token:   os.Getenv("TOKEN"),
	})
}
