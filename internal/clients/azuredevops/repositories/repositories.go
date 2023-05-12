package repositories

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/projects"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

type GitRepository struct {
	Id            *string               `json:"id,omitempty"`
	Name          *string               `json:"name,omitempty"`
	Project       *projects.TeamProject `json:"project,omitempty"`
	DefaultBranch *string               `json:"defaultBranch,omitempty"`
	RemoteUrl     *string               `json:"remoteUrl,omitempty"`
	SshUrl        *string               `json:"sshUrl,omitempty"`
	Url           *string               `json:"url,omitempty"`
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
	Date *azuredevops.Time `json:"date,omitempty"`
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
	Date       *azuredevops.Time `json:"date,omitempty"`
	PushId     *int              `json:"pushId,omitempty"`
	Url        *string           `json:"url,omitempty"`
	Commits    *[]GitCommitRef   `json:"commits,omitempty"`
	RefUpdates *[]GitRefUpdate   `json:"refUpdates,omitempty"`
	Repository *GitRepository    `json:"repository,omitempty"`
}

type GetOptions struct {
	Organization string
	Project      string
	Repository   string
}

// Get retrieve a git repository.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*GitRepository, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories", opts.Repository),
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
	val := &GitRepository{}

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
	ProjectId    string
	Name         string
}

// Create creates a git repository in a team project.
// POST https://dev.azure.com/{organization}/{project}/_apis/git/repositories?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*GitRepository, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(&GitRepository{
		Name: &opts.Name,
		Project: &projects.TeamProject{
			Id: helpers.StringPtr(opts.ProjectId),
		},
	}))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &GitRepository{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusCreated),
		},
	})
	return val, err
}

type DeleteOptions struct {
	Organization string
	Project      string
	RepositoryId string
}

// Delete a git repository.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories/", opts.RepositoryId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return err
	}
	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	return httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		Validators: []httplib.HandleResponseFunc{
			httplib.CheckStatus(http.StatusOK, http.StatusNoContent),
		},
	})
}

// Destroy (hard delete) a soft-deleted Git repository.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/git/recycleBin/repositories/{repositoryId}?api-version=7.0
func DeleteFromRecycleBin(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/recycleBin/repositories", opts.RepositoryId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	return httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		Validators: []httplib.HandleResponseFunc{
			httplib.CheckStatus(http.StatusOK, http.StatusNoContent),
		},
	})
}

type ListResponseValue struct {
	Count int              `json:"count"`
	Value []*GitRepository `json:"value,omitempty"`
}

type ListOptions struct {
	Organization  string
	Project       string
	IncludeHidden bool
}

// List all repositorires in the organization that the authenticated user has access to.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListResponseValue, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	if opts.IncludeHidden {
		params = append(params, "includeHidden", strconv.FormatBool(opts.IncludeHidden))
	}

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
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

	apiErr := &azuredevops.APIError{}
	val := &ListResponseValue{}

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

type FindOptions struct {
	Organization string
	Project      string
	Name         string
}

func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*GitRepository, error) {
	all, err := List(ctx, cli, ListOptions{
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
func CreatePush(ctx context.Context, cli *azuredevops.Client, opts GitPushOptions) (*GitPush, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/git/repositories", opts.RepositoryId, "pushes"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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

	apiErr := &azuredevops.APIError{}
	val := &GitPush{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod:      cli.AuthMethod(),
		Verbose:         cli.Verbose(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK, http.StatusCreated),
		},
	})
	return val, err
}
