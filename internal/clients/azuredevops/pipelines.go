package azuredevops

import (
	"context"
	"net/http"
	"path"

	"github.com/lucasepe/httplib"
)

type ConfigurationType string

const (
	ConfigurationUnknown            ConfigurationType = "unknown"
	ConfigurationYaml               ConfigurationType = "yaml"
	ConfigurationDesignerJson       ConfigurationType = "designerJson"
	ConfigurationJustInTime         ConfigurationType = "justInTime"
	ConfigurationDesignerHyphenJson ConfigurationType = "designerHyphenJson"
)

type BuildRepositoryType string

const (
	BuildRepositoryUnknown                 BuildRepositoryType = "unknown"
	BuildRepositoryGitHub                  BuildRepositoryType = "gitHub"
	BuildRepositoryAzureReposGit           BuildRepositoryType = "azureReposGit"
	BuildRepositoryAzureReposGitHyphenated BuildRepositoryType = "azureReposGit"
)

type BuildRepository struct {
	//The ID of the repository.
	Id string `json:"id,omitempty"`
	//The friendly name of the repository.
	Name string `json:"name,omitempty"`
	// The type of the repository.
	Type BuildRepositoryType `json:"type,omitempty"`
}

type PipelineConfiguration struct {
	// Type of configuration.
	Type ConfigurationType `json:"type,omitempty"`

	//The folder path of the definition.
	Path *string `json:"path,omitempty"`

	Repository *BuildRepository `json:"repository,omitempty"`
}

// Pipeline define a pipeline.
type Pipeline struct {
	// Pipeline ID
	Id *int `json:"id,omitempty"`
	// Pipeline folder
	Folder string `json:"folder,omitempty"`
	// Pipeline name
	Name string `json:"name,omitempty"`
	// Configuration parameters of the pipeline.
	Configuration *PipelineConfiguration `json:"configuration,omitempty"`
	// Revision number
	Revision *int `json:"revision,omitempty"`
	// URL of the pipeline
	Url *string `json:"url,omitempty"`
}

// Arguments for the CreatePipeline function
type CreatePipelineOptions struct {
	Organization string
	Project      string
	Pipeline     Pipeline
}

// CreatePipeline creates a pipeline.
// POST https://dev.azure.com/{organization}/{project}/_apis/pipelines?api-version=7.0
func (c *Client) CreatePipeline(ctx context.Context, opts CreatePipelineOptions) (*Pipeline, error) {
	ub := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines"),
		Params:  []string{apiVersionKey, apiVersionVal},
	})

	req, err := httplib.NewPostRequest(ub, httplib.ToJSON(&opts.Pipeline))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &Pipeline{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	return val, err
}
