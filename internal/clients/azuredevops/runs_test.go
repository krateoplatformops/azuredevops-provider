//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/lucasepe/httplib"
)

func TestRunPipeline(t *testing.T) {
	cli := createAzureDevopsClient()

	pipelineId, err := strconv.Atoi(os.Getenv("PIPELINE_ID"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := cli.RunPipeline(context.TODO(), RunPipelineOptions{
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
	res, err := cli.GetRun(context.TODO(), GetRunOptions{
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
