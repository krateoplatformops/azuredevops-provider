package pullrequests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"time"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
)

// Repository represents the repository object.
type Repository struct {
	ID              string  `json:"id,omitempty"`
	Name            string  `json:"name,omitempty"`
	URL             string  `json:"url,omitempty"`
	Project         Project `json:"project,omitempty"`
	Size            int     `json:"size,omitempty"`
	RemoteURL       string  `json:"remoteUrl,omitempty"`
	SshURL          string  `json:"sshUrl,omitempty"`
	WebURL          string  `json:"webUrl,omitempty"`
	IsDisabled      bool    `json:"isDisabled,omitempty"`
	IsInMaintenance bool    `json:"isInMaintenance,omitempty"`
}

// Project represents the project object.
type Project struct {
	ID             string    `json:"id,omitempty"`
	Name           string    `json:"name,omitempty"`
	URL            string    `json:"url,omitempty"`
	State          string    `json:"state,omitempty"`
	Revision       int       `json:"revision,omitempty"`
	Visibility     string    `json:"visibility,omitempty"`
	LastUpdateTime time.Time `json:"lastUpdateTime,omitempty"`
}

// CreatedBy represents the createdBy object.
type CreatedBy struct {
	DisplayName string `json:"displayName,omitempty"`
	URL         string `json:"url,omitempty"`
	Links       struct {
		Avatar struct {
			Href string `json:"href,omitempty"`
		} `json:"avatar,omitempty"`
	} `json:"_links,omitempty"`
	ID         string `json:"id,omitempty"`
	UniqueName string `json:"uniqueName,omitempty"`
	ImageURL   string `json:"imageUrl,omitempty"`
	Descriptor string `json:"descriptor,omitempty"`
}

// LastMergeSourceCommit represents the last merge source commit object.
type LastMergeSourceCommit struct {
	CommitID string `json:"commitId,omitempty"`
	URL      string `json:"url,omitempty"`
}

// LastMergeTargetCommit represents the last merge target commit object.
type LastMergeTargetCommit struct {
	CommitID string `json:"commitId,omitempty"`
	URL      string `json:"url,omitempty"`
}

type IdentityRef struct {
	Id          string `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UniqueName  string `json:"uniqueName,omitempty"`
	Url         string `json:"url,omitempty"`
	ImageUrl    string `json:"imageUrl,omitempty"`
	Descriptor  string `json:"descriptor,omitempty"`
}
type GitPullRequestMergeOptions struct {
	SquashMerge        bool   `json:"squashMerge,omitempty"`
	CreateMergeCommit  bool   `json:"createMergeCommit,omitempty"`
	MergeCommitMessage string `json:"mergeCommitMessage,omitempty"`
	MergeStrategy      string `json:"mergeStrategy,omitempty"`
}
type CompletionOptions struct {
	BypassPolicy                bool   `json:"bypassPolicy,omitempty"`
	BypassReason                string `json:"bypassReason,omitempty"`
	DeleteSourceBranch          bool   `json:"deleteSourceBranch,omitempty"`
	MergeCommitMessage          string `json:"mergeCommitMessage,omitempty"`
	MergeStrategy               string `json:"mergeStrategy,omitempty"`
	SquashMerge                 bool   `json:"squaredMerge,omitempty"`
	TransitionWorkItems         bool   `json:"transitionWorkItems,omitempty"`
	TriggeredByAutoComplete     bool   `json:"triggeredByAutoComplete,omitempty"`
	AutoCompleteIgnoreConfigIds []int  `json:"autoCompleteIgnoreConfigIds,omitempty"`
}

// PullRequest represents the pull request object.
type PullRequest struct {
	AutoCompleteSetBy     *IdentityRef                `json:"autoCompleteSetBy,omitempty"`
	PullRequestId         int                         `json:"pullRequestId,omitempty"`
	CodeReviewId          int                         `json:"codeReviewId,omitempty"`
	Status                string                      `json:"status,omitempty"`
	CreatedBy             *CreatedBy                  `json:"createdBy,omitempty"`
	CreationDate          interface{}                 `json:"creationDate,omitempty"`
	Title                 string                      `json:"title,omitempty"`
	Description           string                      `json:"description,omitempty"`
	SourceRefName         string                      `json:"sourceRefName,omitempty"`
	TargetRefName         string                      `json:"targetRefName,omitempty"`
	MergeStatus           string                      `json:"mergeStatus,omitempty"`
	IsDraft               bool                        `json:"isDraft,omitempty"`
	MergeID               string                      `json:"mergeId,omitempty"`
	LastMergeSourceCommit *LastMergeSourceCommit      `json:"lastMergeSourceCommit,omitempty"`
	LastMergeTargetCommit *LastMergeTargetCommit      `json:"lastMergeTargetCommit,omitempty"`
	Reviewers             []string                    `json:"reviewers,omitempty"`
	URL                   string                      `json:"url,omitempty"`
	Links                 interface{}                 `json:"_links,omitempty"`
	SupportsIterations    bool                        `json:"supportsIterations,omitempty"`
	ArtifactID            string                      `json:"artifactId,omitempty"`
	CompletionOptions     *CompletionOptions          `json:"completionOptions,omitempty"`
	MergeOptions          *GitPullRequestMergeOptions `json:"mergeOptions,omitempty"`
}

type ListOptions struct {
	Organization string
	ProjectId    string
	RepositoryId string
}

type ListProjectsResponseValue struct {
	Value []*PullRequest `json:"value"`
	Count int            `json:"count"`
}

// Retrieve all pull requests matching a specified criteria.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}/pullrequests?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListProjectsResponseValue, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
			Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories", opts.RepositoryId, "pullrequests"),
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
	val := &ListProjectsResponseValue{
		Value: []*PullRequest{},
	}

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
			return nil
		},
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type GetOptions struct {
	Organization  string
	ProjectId     string
	RepositoryId  string
	PullRequestId string
}

// Retrieve a pull request.
// GET https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}/pullrequests/{pullRequestId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*PullRequest, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories", opts.RepositoryId, "pullrequests", opts.PullRequestId),
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
	val := &PullRequest{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if val != nil && reflect.DeepEqual(*val, PullRequest{}) {
		return nil, err
	}

	return val, err
}

type CreateOptions struct {
	Organization string
	ProjectId    string
	RepositoryId string
	PullRequest  *PullRequest
}

// Create a pull request.
// POST https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}/pullrequests?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*PullRequest, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories", opts.RepositoryId, "pullrequests"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.PullRequest))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &PullRequest{}
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
	Organization  string
	ProjectId     string
	RepositoryId  string
	PullRequestId string
	PullRequest   *PullRequest
}

// Update a pull request
// PATCH https://dev.azure.com/{organization}/{project}/_apis/git/repositories/{repositoryId}/pullrequests/{pullRequestId}?api-version=7.0
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*PullRequest, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/git/repositories", opts.RepositoryId, "pullrequests", opts.PullRequestId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	pr := PullRequest{
		Status:            opts.PullRequest.Status,
		Title:             opts.PullRequest.Title,
		Description:       opts.PullRequest.Description,
		TargetRefName:     opts.PullRequest.TargetRefName,
		CompletionOptions: opts.PullRequest.CompletionOptions,
		MergeOptions:      opts.PullRequest.MergeOptions,
		AutoCompleteSetBy: opts.PullRequest.AutoCompleteSetBy,
	}

	req, err := httplib.Patch(uri.String(), httplib.ToJSON(pr))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &PullRequest{}
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

// Arguments for the FindProjects function
type FindOptions struct {
	Organization  string
	ProjectId     string
	RepositoryId  string
	Title         string
	SourceRefName string
	TargetRefName string
}

// Find utility method to look for a specific project.
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*PullRequest, error) {
	projects, err := List(ctx, cli, ListOptions{
		Organization: opts.Organization,
		ProjectId:    opts.ProjectId,
		RepositoryId: opts.RepositoryId,
	})
	if err != nil {
		return nil, err
	}

	for _, project := range projects.Value {
		if project.Title == opts.Title && project.SourceRefName == opts.SourceRefName && project.TargetRefName == opts.TargetRefName {
			return project, nil
		}
	}

	return nil, &httplib.StatusError{
		StatusCode: http.StatusNotFound,
		Inner:      fmt.Errorf("pull request not found - title: '%s' - sourceRefName: '%s' - targetRefName: '%s'", opts.Title, opts.SourceRefName, opts.TargetRefName),
	}
}
