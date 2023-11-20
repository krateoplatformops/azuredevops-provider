package pipelines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
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

// Options for the Create function
type CreateOptions struct {
	Organization string
	Project      string
	Pipeline     Pipeline
}

// Create creates a pipeline.
// POST https://dev.azure.com/{organization}/{project}/_apis/pipelines?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*Pipeline, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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

	apiErr := &azuredevops.APIError{}
	val := &Pipeline{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	return val, err
}

type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	Project string
	// (required) The pipeline ID
	PipelineId string
	// (optional) The pipeline version
	PipelineVersion *string
}

// Get gets a pipeline, optionally at the specified version
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines/{pipelineId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*Pipeline, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines", opts.PipelineId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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

	apiErr := &azuredevops.APIError{}
	val := &Pipeline{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if val != nil && reflect.DeepEqual(*val, Pipeline{}) {
		return nil, err
	}

	return val, err
}

type ListOptions struct {
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

// List get a list of pipelines.
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListPipelinesResponseValue, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
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
			BaseURL: cli.BaseURL(azuredevops.Default),
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

	apiErr := &azuredevops.APIError{}
	val := &ListPipelinesResponseValue{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
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

type FindOptions struct {
	Organization string
	Project      string
	Name         string
}

// Find utility method to look for a specific pipeline.
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*Pipeline, error) {
	var continutationToken string
	for {
		top := int(30)
		res, err := List(ctx, cli, ListOptions{
			Organization:      opts.Organization,
			Project:           opts.Project,
			Top:               &top,
			ContinuationToken: &continutationToken,
		})
		if err != nil {
			return nil, err
		}

		for _, el := range res.Value {
			if strings.EqualFold(el.Name, opts.Name) {
				return &el, nil
			}
		}

		continutationToken = *res.ContinuationToken
		if continutationToken == "" {
			break
		}
	}

	return nil, &httplib.StatusError{
		StatusCode: http.StatusNotFound,
		Inner:      fmt.Errorf("pipeline '%s' not found", opts.Name),
	}
}
