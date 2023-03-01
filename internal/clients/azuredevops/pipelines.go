package azuredevops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
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
	BuildRepositoryAzureReposGitHyphenated BuildRepositoryType = "azureReposGitHyphenated"
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
	Id *int32 `json:"id,omitempty"`
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
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines"),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(&opts.Pipeline))
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

type GetPipelineOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	Project string
	// (required) The pipeline ID
	PipelineId string
	// (optional) The pipeline version
	PipelineVersion *string
}

// GetPipeline gets a pipeline, optionally at the specified version
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines/{pipelineId}?api-version=7.0
func (c *Client) GetPipeline(ctx context.Context, opts GetPipelineOptions) (*Pipeline, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines", opts.PipelineId),
		Params:  []string{apiVersionKey, apiVersionVal},
	}
	if opts.PipelineVersion != nil {
		ubo.Params = append(ubo.Params, "pipelineVersion", helpers.String(opts.PipelineVersion))
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &Pipeline{}

	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		Verbose:         c.verbose,
		AuthMethod:      c.authMethod,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type ListPipelinesOptions struct {
	Organization string
	Project      string
	// (optional)
	OrderBy *string
	// (optional)
	Top *int
	// (optional)
	Skip *int
	// (optional)
	ContinuationToken *string
}

type ListPipelinesResponseValue struct {
	Count             int        `json:"count"`
	Value             []Pipeline `json:"value,omitempty"`
	ContinuationToken *string    `json:"continuationToken,omitempty"`
}

// Get a list of pipelines.
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines?api-version=7.0
func (c *Client) ListPipelines(ctx context.Context, opts ListPipelinesOptions) (*ListPipelinesResponseValue, error) {
	params := []string{apiVersionKey, apiVersionVal}
	if opts.OrderBy != nil {
		params = append(params, "orderBy", string(*opts.OrderBy))
	}
	if opts.Top != nil {
		params = append(params, "$top", strconv.Itoa(*opts.Top))
	}
	if opts.Skip != nil {
		params = append(params, "$skip", strconv.Itoa(*opts.Skip))
	}
	if opts.ContinuationToken != nil {
		params = append(params, "continuationToken", *opts.ContinuationToken)
	}

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: c.baseURL,
			Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines"),
			Params:  params,
		}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Get(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &ListPipelinesResponseValue{}

	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod: c.authMethod,
		Verbose:    c.verbose,
		ResponseHandler: func(res *http.Response) error {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			if err = json.Unmarshal(data, val); err != nil {
				return err
			}

			val.ContinuationToken = helpers.StringPtr(res.Header.Get("X-Ms-Continuationtoken"))
			return nil
		},
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}
