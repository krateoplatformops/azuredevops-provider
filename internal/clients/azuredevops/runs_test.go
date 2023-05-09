//go:build integration
// +build integration

package azuredevops

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
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
