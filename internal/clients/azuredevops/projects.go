package azuredevops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"

	"gihtub.com/krateoplatformops/azuredevops-provider/internal/httplib"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
)

type ProjectState string

// ProjectState types.
const (
	// Project is in the process of being deleted.
	StateDeleting ProjectState = "deleting"
	// Project is in the process of being created.
	StateNew ProjectState = "new"
	// Project is completely created and ready to use.
	StateWellFormed ProjectState = "wellFormed"
	// Project has been queued for creation, but the process has not yet started.
	StateCreatePending ProjectState = "createPending"
	// All projects regardless of state.
	StateAll ProjectState = "all"
	// Project has not been changed.
	StateUnchanged ProjectState = "unchanged"
	// Project has been deleted.
	StateDeleted ProjectState = "deleted"
)

type ProjectVisibility string

const (
	VisibilityPrivate ProjectVisibility = "private"
	VisibilityPublic  ProjectVisibility = "public"
)

type WebApiTeamRef struct {
	// Team (Identity) Guid. A Team Foundation ID.
	Id *string `json:"id,omitempty"`
	// Team name
	Name *string `json:"name,omitempty"`
	// Team REST API Url
	Url *string `json:"url,omitempty"`
}

// Represents a Team Project object.
type TeamProject struct {
	// Project abbreviation.
	Abbreviation *string `json:"abbreviation,omitempty"`
	// Url to default team identity image.
	DefaultTeamImageUrl *string `json:"defaultTeamImageUrl,omitempty"`
	// The project's description (if any).
	Description *string `json:"description,omitempty"`
	// Project identifier.
	Id *string `json:"id,omitempty"`
	// Project last update time.
	LastUpdateTime *Time `json:"lastUpdateTime,omitempty"`
	// Project name.
	Name *string `json:"name,omitempty"`
	// Project revision.
	Revision *uint64 `json:"revision,omitempty"`
	// Project state.
	State *ProjectState `json:"state,omitempty"`
	// Url to the full version of the object.
	Url *string `json:"url,omitempty"`
	// Project visibility.
	Visibility *ProjectVisibility `json:"visibility,omitempty"`
	// The links to other objects related to this object.
	Links any `json:"_links,omitempty"`
	// Set of capabilities this project has (such as process template & version control).
	Capabilities *map[string]map[string]string `json:"capabilities,omitempty"`
	// The shallow ref to the default team.
	DefaultTeam *WebApiTeamRef `json:"defaultTeam,omitempty"`
}

// Arguments for the ListProjects function
type ListProjectsOpts struct {
	Organization string
	// (optional) Filter on team projects in a specific team project state (default: WellFormed).
	StateFilter *ProjectState
	// (optional)
	Top *int
	// (optional)
	Skip *int
	// (optional)
	ContinuationToken *string
}

// Return type for the GetProjects function
type ListProjectsResponseValue struct {
	Count             int           `json:"count"`
	Value             []TeamProject `json:"value,omitempty"`
	ContinuationToken *string       `json:"continuationToken,omitempty"`
}

// Get all projects in the organization that the authenticated user has access to.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/core/projects/list?view=azure-devops-rest-7.0&tabs=HTTP#teamprojectreference
func ListProjects(ctx context.Context, cli *Client, opts ListProjectsOpts) (*ListProjectsResponseValue, error) {
	apiPath := path.Join(opts.Organization, "_apis/projects")
	queryParams := map[string]string{}
	if opts.StateFilter != nil {
		queryParams["stateFilter"] = string(*opts.StateFilter)
	}
	if opts.Top != nil {
		queryParams["$top"] = strconv.Itoa(*opts.Top)
	}
	if opts.Skip != nil {
		queryParams["$skip"] = strconv.Itoa(*opts.Skip)
	}
	if opts.ContinuationToken != nil {
		queryParams["continuationToken"] = *opts.ContinuationToken
	}

	req, err := cli.newGetRequest(apiPath, queryParams)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &ListProjectsResponseValue{}

	err = httplib.Call(cli.httpClient, req, httplib.CallOpts{
		Verbose: cli.options.Verbose,
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
		Validators: []httplib.ResponseHandler{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type GetProjectOpts struct {
	Organization string
	ProjectId    string
}

// Get project with the specified id or name, optionally including capabilities.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/core/projects/get?view=azure-devops-rest-7.0
func GetProject(ctx context.Context, cli *Client, opts GetProjectOpts) (*TeamProject, error) {
	apiPath := path.Join(opts.Organization, "_apis/projects", opts.ProjectId)
	req, err := cli.newGetRequest(apiPath, nil)
	if err != nil {
		return nil, err
	}

	apiErr := &APIError{}
	val := &TeamProject{}

	err = httplib.Call(cli.httpClient, req, httplib.CallOpts{
		Verbose:         cli.options.Verbose,
		ResponseHandler: httplib.ToJSON(val),
		Validators: []httplib.ResponseHandler{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type CreateProjectOpts struct {
	Organization string
	TeamProject  *TeamProject
}

// Queues a project to be created. Use the GetOperation to periodically check for create project status.
// POST https://dev.azure.com/{organization}/_apis/projects?api-version=7.0
func CreateProject(ctx context.Context, cli *Client, opts CreateProjectOpts) (*OperationReference, error) {
	apiPath := path.Join(opts.Organization, "_apis/projects")
	req, err := cli.newPostRequest(apiPath, nil, opts.TeamProject)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &OperationReference{}
	err = httplib.Call(cli.httpClient, req, httplib.CallOpts{
		Verbose:         cli.options.Verbose,
		ResponseHandler: httplib.ToJSON(val),
		Validators: []httplib.ResponseHandler{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	return val, err
}

type DeleteProjectOpts struct {
	Organization string
	ProjectId    string
}

// Queues a project to be deleted. Use the GetOperation to periodically check for delete project status.
// DELETE https://dev.azure.com/{organization}/_apis/projects/{projectId}?api-version=7.0
func DeleteProject(ctx context.Context, cli *Client, opts DeleteProjectOpts) (*OperationReference, error) {
	apiPath := path.Join(opts.Organization, "_apis/projects/", opts.ProjectId)
	req, err := cli.newDeleteRequest(apiPath, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &OperationReference{}
	err = httplib.Call(cli.httpClient, req, httplib.CallOpts{
		Verbose:         cli.options.Verbose,
		ResponseHandler: httplib.ToJSON(val),
		Validators: []httplib.ResponseHandler{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	return val, err
}
