package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
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

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
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
	Visibility Visibility `json:"visibility,omitempty"`

	// Set of capabilities this project has (such as process template & version control).
	Capabilities *Capabilities `json:"capabilities,omitempty"`

	// Project revision.
	Revision *uint64 `json:"revision,omitempty"`

	// Project state.
	State *ProjectState `json:"state,omitempty"`
}

// Options for the List function
type ListOptions struct {
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
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListProjectsResponseValue, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
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
			BaseURL: cli.BaseURL(azuredevops.Default),
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

	apiErr := &azuredevops.APIError{}
	val := &ListProjectsResponseValue{}

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

type GetOptions struct {
	Organization string
	ProjectId    string
}

// Get project with the specified id or name, optionally including capabilities.
// https://learn.microsoft.com/en-us/rest/api/azure/devops/core/projects/get?view=azure-devops-rest-7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*TeamProject, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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
	val := &TeamProject{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type CreateOptions struct {
	Organization string
	TeamProject  *TeamProject
}

// Queues a project to be created. Use the GetOperation to periodically check for create project status.
// POST https://dev.azure.com/{organization}/_apis/projects?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*azuredevops.OperationReference, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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

	apiErr := &azuredevops.APIError{}
	val := &azuredevops.OperationReference{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusAccepted),
		},
	})
	return val, err
}

type UpdateOptions struct {
	Organization string
	ProjectId    string
	TeamProject  *TeamProject
}

// Update an existing project's name, abbreviation, description, or restore a project.
// PATCH https://dev.azure.com/{organization}/_apis/projects/{projectId}?api-version=7.0
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*azuredevops.OperationReference, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects", opts.ProjectId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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

	apiErr := &azuredevops.APIError{}
	val := &azuredevops.OperationReference{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusAccepted),
		},
	})
	return val, err
}

type DeleteOptions struct {
	Organization string
	ProjectId    string
}

// Queues a project to be deleted. Use the GetOperation to periodically check for delete project status.
// DELETE https://dev.azure.com/{organization}/_apis/projects/{projectId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) (*azuredevops.OperationReference, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/projects/", opts.ProjectId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &azuredevops.OperationReference{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusAccepted),
		},
	})
	return val, err
}

// Arguments for the FindProjects function
type FindOptions struct {
	Organization string
	Name         string
}

// Find utility method to look for a specific project.
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*TeamProject, error) {
	var continutationToken string
	for {
		top := int(30)
		//filter := StateWellFormed
		res, err := List(ctx, cli, ListOptions{
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
