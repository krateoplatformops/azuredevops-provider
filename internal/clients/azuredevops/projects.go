package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
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

type Versioncontrol struct {
	// SourceControlType:
	SourceControlType string `json:"sourceControlType,omitempty"`
}

// ProcessTemplate define reusable content in Azure Devops.
type ProcessTemplate struct {
	// TemplateTypeId: id of the desired template
	TemplateTypeId string `json:"templateTypeId,omitempty"`
}

// Capabilities this project has
type Capabilities struct {
	Versioncontrol *Versioncontrol `json:"versioncontrol,omitempty"`

	ProcessTemplate *ProcessTemplate `json:"processTemplate,omitempty"`
}

// Represents a Team Project object.
type TeamProject struct {
	// Project identifier.
	Id *string `json:"id,omitempty"`

	// Project name.
	Name string `json:"name,omitempty"`

	// The project's description (if any).
	Description *string `json:"description,omitempty"`

	// Project visibility.
	Visibility ProjectVisibility `json:"visibility,omitempty"`

	// Set of capabilities this project has (such as process template & version control).
	Capabilities *Capabilities `json:"capabilities,omitempty"`

	// Project revision.
	Revision *uint64 `json:"revision,omitempty"`

	// Project state.
	State *ProjectState `json:"state,omitempty"`
}

// Arguments for the ListProjects function
type ListProjectsOptions struct {
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
	Count             int            `json:"count"`
	Value             []*TeamProject `json:"value,omitempty"`
	ContinuationToken *string        `json:"continuationToken,omitempty"`
}

// Get all projects in the organization that the authenticated user has access to.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/core/projects/list?view=azure-devops-rest-7.0&tabs=HTTP#teamprojectreference
func (c *Client) ListProjects(ctx context.Context, opts ListProjectsOptions) (*ListProjectsResponseValue, error) {
	params := []string{apiVersionKey, apiVersionVal}
	if opts.StateFilter != nil {
		params = append(params, "stateFilter", string(*opts.StateFilter))
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
			Path:    path.Join(opts.Organization, "_apis/projects"),
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
	val := &ListProjectsResponseValue{}

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

type GetProjectOptions struct {
	Organization string
	ProjectId    string
}

// Get project with the specified id or name, optionally including capabilities.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/core/projects/get?view=azure-devops-rest-7.0
func (c *Client) GetProject(ctx context.Context, opts GetProjectOptions) (*TeamProject, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectId),
		Params:  []string{apiVersionKey, apiVersionVal},
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
	val := &TeamProject{}

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

type CreateProjectOptions struct {
	Organization string
	TeamProject  *TeamProject
}

// Queues a project to be created. Use the GetOperation to periodically check for create project status.
// POST https://dev.azure.com/{organization}/_apis/projects?api-version=7.0
func (c *Client) CreateProject(ctx context.Context, opts CreateProjectOptions) (*OperationReference, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, "_apis/projects"),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.TeamProject))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &OperationReference{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusAccepted),
		},
	})
	return val, err
}

type UpdateProjectOptions struct {
	Organization string
	ProjectId    string
	TeamProject  *TeamProject
}

// Update an existing project's name, abbreviation, description, or restore a project.
// PATCH https://dev.azure.com/{organization}/_apis/projects/{projectId}?api-version=7.0
func (c *Client) UpdateProject(ctx context.Context, opts UpdateProjectOptions) (*OperationReference, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectId),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Patch(uri.String(), httplib.ToJSON(opts.TeamProject))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &OperationReference{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusAccepted),
		},
	})
	return val, err
}

type DeleteProjectOptions struct {
	Organization string
	ProjectId    string
}

// Queues a project to be deleted. Use the GetOperation to periodically check for delete project status.
// DELETE https://dev.azure.com/{organization}/_apis/projects/{projectId}?api-version=7.0
func (c *Client) DeleteProject(ctx context.Context, opts DeleteProjectOptions) (*OperationReference, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, "_apis/projects/", opts.ProjectId),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &OperationReference{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusAccepted),
		},
	})
	return val, err
}

// Arguments for the FindProjects function
type FindProjectsOptions struct {
	Organization string
	Name         string
}

// FindProject utility method to look for a specific project.
func (c *Client) FindProject(ctx context.Context, opts FindProjectsOptions) (*TeamProject, error) {
	var continutationToken string
	for {
		top := int(30)
		//filter := StateWellFormed
		res, err := c.ListProjects(ctx, ListProjectsOptions{
			Organization: opts.Organization,
			//StateFilter:       &filter,
			Top:               &top,
			ContinuationToken: &continutationToken,
		})
		if err != nil {
			return nil, err
		}

		for _, el := range res.Value {
			if strings.EqualFold(el.Name, opts.Name) {
				return el, nil
			}
		}

		continutationToken = *res.ContinuationToken
		if continutationToken == "" {
			break
		}
	}

	return nil, &httplib.StatusError{
		StatusCode: http.StatusNotFound,
		Inner:      fmt.Errorf("project '%s' not found", opts.Name),
	}
}
