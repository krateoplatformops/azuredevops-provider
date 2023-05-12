//go:build integration
// +build integration

package runs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/dotenv"
	"github.com/lucasepe/httplib"
)

func TestRunPipeline(t *testing.T) {
	cli := createAzureDevopsClient()

	pipelineId, err := strconv.Atoi(os.Getenv("PIPELINE_ID"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Run(context.TODO(), cli, RunOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		PipelineId:   pipelineId,
	})
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(res)
}

func TestGetRun(t *testing.T) {
	cli := createAzureDevopsClient()

	pipelineId, err := strconv.Atoi(os.Getenv("PIPELINE_ID"))
	if err != nil {
		t.Fatal(err)
	}

	runId, err := strconv.Atoi(os.Getenv("RUN_ID"))
	if err != nil {
		t.Fatal(err)
	}
	runId = 3
	res, err := Get(context.TODO(), cli, GetOptions{
		Organization: os.Getenv("ORG"),
		Project:      os.Getenv("PROJECT_NAME"),
		PipelineId:   pipelineId,
		RunId:        runId,
	})
	if err != nil && !httplib.IsNotFoundError(err) {
		t.Fatal(err)
	}

	if res != nil && res.Id != nil {
		spew.Dump(res)
		fmt.Println()
		fmt.Println("state: ", *res.State)
	}
}

func createAzureDevopsClient() *azuredevops.Client {
	env, _ := dotenv.FromFile("../../../../.env")
	dotenv.PutInEnv(env, false)

	return azuredevops.NewClient(azuredevops.ClientOptions{
		Verbose: false,
		Token:   os.Getenv("TOKEN"),
	})
}
