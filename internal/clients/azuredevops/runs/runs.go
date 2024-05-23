package runs

import (
	"context"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"

	runsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/runs/v1alpha1"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

// A reference to a Pipeline.
type PipelineReference struct {
	// Pipeline folder
	Folder *string `json:"folder,omitempty"`
	// Pipeline ID
	Id *int `json:"id,omitempty"`
	// Pipeline name
	Name *string `json:"name,omitempty"`
	// Revision number
	Revision *int    `json:"revision,omitempty"`
	Url      *string `json:"url,omitempty"`
}

type Container struct {
	Environment     map[string]string `json:"environment,omitempty"`
	Image           *string           `json:"image,omitempty"`
	MapDockerSocket *bool             `json:"mapDockerSocket,omitempty"`
	Options         *string           `json:"options,omitempty"`
	Ports           []string          `json:"ports,omitempty"`
	Volumes         []string          `json:"volumes,omitempty"`
}

type ContainerResource struct {
	Container *Container `json:"container,omitempty"`
}

type PipelineResource struct {
	Pipeline *PipelineReference `json:"pipeline,omitempty"`
	Version  *string            `json:"version,omitempty"`
}

type RepositoryResource struct {
	RefName    *string     `json:"refName,omitempty"`
	Repository *Repository `json:"repository,omitempty"`
	Version    *string     `json:"version,omitempty"`
}

type RepositoryType string

const (
	Unknown                 RepositoryType = "unknown"
	GitHub                  RepositoryType = "gitHub"
	AzureReposGit           RepositoryType = "azureReposGit"
	GitHubEnterprise        RepositoryType = "gitHubEnterprise"
	AzureReposGitHyphenated RepositoryType = "azureReposGitHyphenated"
)

type Repository struct {
	Type *RepositoryType `json:"type,omitempty"`
}

type RunResources struct {
	Containers   *map[string]ContainerResource  `json:"containers,omitempty"`
	Pipelines    *map[string]PipelineResource   `json:"pipelines,omitempty"`
	Repositories *map[string]RepositoryResource `json:"repositories,omitempty"`
}

type RunInfo struct {
	Id                 *int                   `json:"id,omitempty"`
	Name               *string                `json:"name,omitempty"`
	Links              interface{}            `json:"_links,omitempty"`
	CreatedDate        *azuredevops.Time      `json:"createdDate,omitempty"`
	FinalYaml          *string                `json:"finalYaml,omitempty"`
	FinishedDate       *azuredevops.Time      `json:"finishedDate,omitempty"`
	Pipeline           *PipelineReference     `json:"pipeline,omitempty"`
	Resources          *RunResources          `json:"resources,omitempty"`
	Result             *string                `json:"result,omitempty"`
	State              *string                `json:"state,omitempty"`
	TemplateParameters map[string]interface{} `json:"templateParameters,omitempty"`
	Url                *string                `json:"url,omitempty"`
	Variables          map[string]Variable    `json:"variables,omitempty"`
}

type Variable struct {
	IsSecret *bool   `json:"isSecret,omitempty"`
	Value    *string `json:"value,omitempty"`
}

type BuildResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type ContainerResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type PackageResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type PipelineResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type RepositoryResourceParameters struct {
	RefName *string `json:"refName,omitempty"`
	// This is the security token to use when connecting to the repository.
	Token *string `json:"token,omitempty"`
	// Optional. This is the type of the token given. If not provided, a type of "Bearer" is assumed. Note: Use "Basic" for a PAT token.
	TokenType *string `json:"tokenType,omitempty"`
	Version   *string `json:"version,omitempty"`
}

type RunResourcesParameters struct {
	Builds       map[string]runsv1alpha1.BuildResourceParameters      `json:"builds,omitempty"`
	Containers   map[string]runsv1alpha1.ContainerResourceParameters  `json:"containers,omitempty"`
	Packages     map[string]runsv1alpha1.PackageResourceParameters    `json:"packages,omitempty"`
	Pipelines    map[string]runsv1alpha1.PipelineResourceParameters   `json:"pipelines,omitempty"`
	Repositories map[string]runsv1alpha1.RepositoryResourceParameters `json:"repositories,omitempty"`
}

type RunPipelineParameters struct {
	// If true, don't actually create a new run.
	// Instead, return the final YAML document after parsing templates.
	// +optional
	PreviewRun *bool `json:"previewRun,omitempty"`

	// The resources the run requires.
	// +optional
	Resources *RunResourcesParameters `json:"resources,omitempty"`

	// +optional
	StagesToSkip []string `json:"stagesToSkip,omitempty"`

	// +optional
	TemplateParameters map[string]string `json:"templateParameters,omitempty"`

	// +optional
	Variables map[string]runsv1alpha1.Variable `json:"variables,omitempty"`
	// YamlOverride: If you use the preview run option, you may optionally supply different YAML.
	// This allows you to preview the final YAML document without committing a changed file.
	// +optional
	YamlOverride *string `json:"yamlOverride,omitempty"`
}

// Options for the Run pipeline function
type RunOptions struct {
	// (required) Optional additional parameters for this run.
	RunParameters *RunPipelineParameters

	//  (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	Project string
	// (required) The pipeline ID.
	PipelineId int
	// (optional) The pipeline version.
	PipelineVersion *int
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.Runs
		if apiVersion != nil {
			if strings.EqualFold(*apiVersion, "none") {
				apiVersionParams = nil
				isNone = true
			} else {
				apiVersionParams = []string{azuredevops.ApiVersionKey, helpers.String(apiVersion)}
			}
		}
	}
	return apiVersionParams, isNone
}

// Run run Pipeline.
// POST https://dev.azure.com/{organization}/{project}/_apis/pipelines/{pipelineId}/runs?api-version=7.0
func Run(ctx context.Context, cli *azuredevops.Client, opts RunOptions) (*RunInfo, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines", strconv.Itoa(opts.PipelineId), "runs"),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	if opts.RunParameters == nil {
		opts.RunParameters = &RunPipelineParameters{
			PreviewRun:   helpers.BoolPtr(false),
			StagesToSkip: []string{},
			Resources:    &RunResourcesParameters{},
		}
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.RunParameters))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &RunInfo{}
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

// Options for the Get run function
type GetOptions struct {
	// (required) The name of the Azure DevOps organization.
	Organization string
	// (required) Project ID or project name
	Project string
	// (required) The pipeline id
	PipelineId int
	// (required) The run id
	RunId int
}

// Gets a run for a particular pipeline.
// GET https://dev.azure.com/{organization}/{project}/_apis/pipelines/{pipelineId}/runs/{runId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*RunInfo, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/pipelines", strconv.Itoa(opts.PipelineId), "runs", strconv.Itoa(opts.RunId)),
		Params:  apiVersionParams,
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
	val := &RunInfo{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if err != nil {
		return nil, err
	}
	if val != nil && reflect.DeepEqual(*val, RunInfo{}) {
		return nil, err
	}

	return val, nil
}
