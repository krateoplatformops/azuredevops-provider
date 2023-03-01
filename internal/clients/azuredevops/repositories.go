package azuredevops

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type GitRepository struct {
	Id            *string      `json:"id,omitempty"`
	Name          *string      `json:"name,omitempty"`
	Project       *TeamProject `json:"project,omitempty"`
	DefaultBranch *string      `json:"defaultBranch,omitempty"`
	RemoteUrl     *string      `json:"remoteUrl,omitempty"`
	SshUrl        *string      `json:"sshUrl,omitempty"`
	Url           *string      `json:"url,omitempty"`
}

type GitRefUpdate struct {
	IsLocked     *bool   `json:"isLocked,omitempty"`
	Name         *string `json:"name,omitempty"`
	NewObjectId  *string `json:"newObjectId,omitempty"`
	OldObjectId  *string `json:"oldObjectId,omitempty"`
	RepositoryId *string `json:"repositoryId,omitempty"`
}

// User info and date for Git operations.
type GitUserDate struct {
	// Date of the Git operation.
	Date *Time `json:"date,omitempty"`
	// Email address of the user performing the Git operation.
	Email *string `json:"email,omitempty"`
	// Name of the user performing the Git operation.
	Name *string `json:"name,omitempty"`
}

// Provides properties that describe a Git commit and associated metadata.
type GitCommitRef struct {
	// Author of the commit.
	Author *GitUserDate `json:"author,omitempty"`
	// An enumeration of the changes included with the commit.
	Changes []GitChange `json:"changes,omitempty"`
	// Comment or message of the commit.
	Comment *string `json:"comment,omitempty"`
	// Committer of the commit.
	Committer *GitUserDate `json:"committer,omitempty"`
}

type VersionControlChangeType string

const (
	ChangeTypeNone         VersionControlChangeType = "none"
	ChangeTypeAdd          VersionControlChangeType = "add"
	ChangeTypeEdit         VersionControlChangeType = "edit"
	ChangeTypeEncoding     VersionControlChangeType = "encoding"
	ChangeTypeRename       VersionControlChangeType = "rename"
	ChangeTypeDelete       VersionControlChangeType = "delete"
	ChangeTypeUndelete     VersionControlChangeType = "undelete"
	ChangeTypeBranch       VersionControlChangeType = "branch"
	ChangeTypeMerge        VersionControlChangeType = "merge"
	ChangeTypeLock         VersionControlChangeType = "lock"
	ChangeTypeRollback     VersionControlChangeType = "rollback"
	ChangeTypeSourceRename VersionControlChangeType = "sourceRename"
	ChangeTypeTargetRename VersionControlChangeType = "targetRename"
	ChangeTypeProperty     VersionControlChangeType = "property"
	ChangeTypeAll          VersionControlChangeType = "all"
)

type ItemContentType string

const (
	ContentTypeRawText       ItemContentType = "rawText"
	ContentTypeBase64Encoded ItemContentType = "base64Encoded"
)

type ItemContent struct {
	Content     string          `json:"content,omitempty"`
	ContentType ItemContentType `json:"contentType,omitempty"`
}

type GitChange struct {
	// The type of change that was made to the item.
	ChangeType VersionControlChangeType `json:"changeType,omitempty"`
	// Content of the item after the change.
	NewContent *ItemContent `json:"newContent,omitempty"`

	Item map[string]string `json:"item,omitempty"`
}

type GitPush struct {
	Date       *Time           `json:"date,omitempty"`
	PushId     *int            `json:"pushId,omitempty"`
	Url        *string         `json:"url,omitempty"`
	Commits    *[]GitCommitRef `json:"commits,omitempty"`
	RefUpdates *[]GitRefUpdate `json:"refUpdates,omitempty"`
	Repository *GitRepository  `json:"repository,omitempty"`
}

type GetRepositoryOptions struct {
	Organization string
	Project      string
	Repository   string
}

// GetRepository retrieve a git repository.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}?api-version=7.0
func (c *Client) GetRepository(ctx context.Context, opts GetRepositoryOptions) (*GitRepository, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories", opts.Repository),
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
	val := &GitRepository{}

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

type CreateRepositoryOptions struct {
	Organization string
	ProjectId    string
	Name         string
}

// CreateRepository creates a git repository in a team project.
// POST https://dev.azure.com/{organization}/{project}/_apis/git/repositories?api-version=7.0
func (c *Client) CreateRepository(ctx context.Context, opts CreateRepositoryOptions) (*GitRepository, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories"),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(&GitRepository{
		Name: &opts.Name,
		Project: &TeamProject{
			Id: helpers.StringPtr(opts.ProjectId),
		},
	}))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &GitRepository{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusCreated),
		},
	})
	return val, err
}

type DeleteRepositoryOptions struct {
	Organization string
	Project      string
	RepositoryId string
}

// Delete a git repository.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}?api-version=7.0
func (c *Client) DeleteRepository(ctx context.Context, opts DeleteRepositoryOptions) error {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories/", opts.RepositoryId),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return err
	}
	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	return httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod: c.authMethod,
		Verbose:    c.verbose,
		Validators: []httplib.HandleResponseFunc{
			httplib.CheckStatus(http.StatusOK, http.StatusNoContent),
		},
	})
}

// Destroy (hard delete) a soft-deleted Git repository.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/git/recycleBin/repositories/{repositoryId}?api-version=7.0
func (c *Client) DeleteRepositoryFromRecycleBin(ctx context.Context, opts DeleteRepositoryOptions) error {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/recycleBin/repositories", opts.RepositoryId),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	return httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod: c.authMethod,
		Verbose:    c.verbose,
		Validators: []httplib.HandleResponseFunc{
			httplib.CheckStatus(http.StatusOK, http.StatusNoContent),
		},
	})
}

type ListRepositoriesResponseValue struct {
	Count int              `json:"count"`
	Value []*GitRepository `json:"value,omitempty"`
}

type ListRepositoriesOptions struct {
	Organization  string
	Project       string
	IncludeHidden bool
}

// List all repositorires in the organization that the authenticated user has access to.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories?api-version=7.0
func (c *Client) ListRepositories(ctx context.Context, opts ListRepositoriesOptions) (*ListRepositoriesResponseValue, error) {
	params := []string{apiVersionKey, apiVersionVal}
	if opts.IncludeHidden {
		params = append(params, "includeHidden", strconv.FormatBool(opts.IncludeHidden))
	}

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: c.baseURL,
			Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories"),
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
	val := &ListRepositoriesResponseValue{}

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

type FindRepositoryOptions struct {
	Organization string
	Project      string
	Name         string
}

func (c *Client) FindRepository(ctx context.Context, opts FindRepositoryOptions) (*GitRepository, error) {
	all, err := c.ListRepositories(ctx, ListRepositoriesOptions{
		Organization:  opts.Organization,
		Project:       opts.Project,
		IncludeHidden: true,
	})
	if err != nil {
		return nil, err
	}

	for _, el := range all.Value {
		if helpers.String(el.Name) == opts.Name {
			return el, nil
		}
	}

	return nil, &httplib.StatusError{
		StatusCode: http.StatusNotFound,
		Inner: fmt.Errorf("GitRepository not found (organization: %s, project: %s, name: %s)",
			opts.Organization, opts.Project, opts.Name),
	}
}

// Arguments for the CreatePush function
type GitPushOptions struct {
	// (required)
	Push *GitPush
	// (required) The name or ID of the repository.
	RepositoryId string
	// (required) Project ID or project name
	Project string
	// (required) Organization
	Organization string
}

// Push changes to the repository.
// POST https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}/pushes?api-version=7.0
func (c *Client) CreatePush(ctx context.Context, opts GitPushOptions) (*GitPush, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: c.baseURL,
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories", opts.RepositoryId, "pushes"),
		Params:  []string{apiVersionKey, apiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.Push))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &APIError{}
	val := &GitPush{}
	err = httplib.Fire(c.httpClient, req, httplib.FireOptions{
		AuthMethod:      c.authMethod,
		Verbose:         c.verbose,
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusCreated),
		},
	})
	return val, err
}
