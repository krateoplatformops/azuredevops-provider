package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
	"k8s.io/apimachinery/pkg/api/resource"
)

type PolicyType struct {
	// Display name of the policy type.
	DisplayName string `json:"displayName"`
	// The policy type ID.
	Id string `json:"id"`
	// The URL where the policy type can be retrieved.
	Url string `json:"url"`
}
type Scope struct {
	RefName      string `json:"refName"`
	MatchKind    string `json:"matchKind"`
	RepositoryId string `json:"repositoryId,omitempty"`
}
type PolicySettings struct {
	MinimumApproverCount      int               `json:"minimumApproverCount"`
	CreatorVoteCounts         bool              `json:"creatorVoteCounts"`
	Scope                     []Scope           `json:"scope"`
	BuildDefinitionId         int               `json:"buildDefinitionId"`
	RequiredReviewerIds       []string          `json:"requiredReviewerIds"`
	FileNamePatterns          []string          `json:"fileNamePatterns"`
	AddedFilesOnly            bool              `json:"addedFilesOnly"`
	Message                   string            `json:"message"`
	EnforceConsistentCase     bool              `json:"enforceConsistentCase"`
	MaximumGitBlobSizeInBytes int               `json:"maximumGitBlobSizeInBytes"`
	UseUncompressedSize       bool              `json:"useUncompressedSize"`
	UseSquashMerge            bool              `json:"useSquashMerge"`
	ManualQueueOnly           bool              `json:"manualQueueOnly"`
	QueueOnSourceUpdateOnly   bool              `json:"queueOnSourceUpdateOnly"`
	DisplayName               string            `json:"displayName"`
	ValidDuration             resource.Quantity `json:"validDuration"`
}

// Policy defines the desired state of Policy
type PolicyBody struct {
	// Type - The policy configuration type.
	Type PolicyType

	// IsBlocking - Indicates whether the policy is blocking.
	IsBlocking bool

	// IsEnabled - Indicates whether the policy is enabled.
	IsEnabled bool

	// IsEnterpriseManaged - If set, this policy requires "Manage Enterprise Policies" permission to create, edit, or delete.
	IsEnterpriseManaged bool

	// URL - The URL where the policy configuration can be retrieved.
	URL string

	// Revision - The policy configuration revision ID.
	Revision int

	// IsDeleted - Indicates whether the policy has been (soft) deleted.
	IsDeleted bool

	// Settings - The policy configuration settings.
	Settings PolicySettings

	// ID - The policy configuration ID.
	ID int
}

func getAPIVersion(cli *azuredevops.Client) (apiVersionParams []string, isNone bool) {
	if cli.ApiVersionConfig != nil {
		apiVersion := cli.ApiVersionConfig.Policies
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

type ListOptions struct {
	Organization string
	ProjectId    string

	Scope             *string
	Top               *int
	ContinuationToken *string
	PolicyType        *string
}

type ListPolicies struct {
	Count             int           `json:"count"`
	Value             []*PolicyBody `json:"value"`
	ContinuationToken *string
}

// Get a list of policy configurations in a project.
// The 'scope' parameter for this API should not be used, except for legacy compatability reasons. It returns specifically scoped policies and does not support heirarchical nesting. Instead, use the /_apis/git/policy/configurations API, which provides first class scope filtering support.
// The optional policyType parameter can be used to filter the set of policies returned from this method.
// GET https://dev.azure.com/{organization}/{project}/_apis/policy/configurations?scope={scope}&$top={$top}&continuationToken={continuationToken}&policyType={policyType}&api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) (*ListPolicies, error) {
	queryParams := []string{}
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	if opts.Scope != nil {
		queryParams = append(queryParams, "scope", *opts.Scope)
	}
	if opts.Top != nil {
		queryParams = append(queryParams, "$top", fmt.Sprintf("%d", *opts.Top))
	}
	if opts.ContinuationToken != nil {
		queryParams = append(queryParams, "continuationToken", *opts.ContinuationToken)
	}
	if opts.PolicyType != nil {
		queryParams = append(queryParams, "policyType", *opts.PolicyType)
	}
	queryParams = append(queryParams, apiVersionParams...)

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
			Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/policy/configurations"),
			Params:  queryParams,
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
	val := &ListPolicies{
		Value: []*PolicyBody{},
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
	// ConfigurationId - The ID of the policy configuration.
	ConfigurationId int
}

// Get a policy configuration by its ID.
// GET https://dev.azure.com/{organization}/{project}/_apis/policy/configurations/{configurationId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*PolicyBody, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/policy/configurations", fmt.Sprint(opts.ConfigurationId)),
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
	val := &PolicyBody{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if reflect.DeepEqual(*val, PolicyBody{}) {
		return nil, err
	}

	return val, err
}

type CreateOptions struct {
	Organization string
	ProjectId    string
	PolicyBody   *PolicyBody
}

// POST https://dev.azure.com/{organization}/{project}/_apis/policy/configurations?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*PolicyBody, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/policy/configurations"),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.PolicyBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &PolicyBody{}
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

type UpdateOptions struct {
	Organization    string
	ProjectId       string
	ConfigurationId int
	PolicyBody      *PolicyBody
}

// Update a policy configuration by its ID.
// PUT https://dev.azure.com/{organization}/{project}/_apis/policy/configurations/{configurationId}?api-version=7.0
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*PolicyBody, error) {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/policy/configurations", fmt.Sprint(opts.ConfigurationId)),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Put(uri.String(), httplib.ToJSON(opts.PolicyBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &PolicyBody{}
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
	Organization    string
	ProjectId       string
	ConfigurationId int
}

// Delete a policy configuration by its ID.
// DELETE https://dev.azure.com/{organization}/{project}/_apis/policy/configurations/{configurationId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	apiVersionParams, isNone := getAPIVersion(cli)
	if len(apiVersionParams) == 0 && !isNone {
		apiVersionParams = []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.ProjectId, "_apis/policy/configurations", fmt.Sprint(opts.ConfigurationId)),
		Params:  apiVersionParams,
	}).Build()
	if err != nil {
		return err
	}

	req, err := httplib.Delete(uri.String())
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:    cli.Verbose(),
		AuthMethod: cli.AuthMethod(),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusNoContent),
		},
	})
	return err
}

// Arguments for the FindProjects function
type FindOptions struct {
	Organization string
	ProjectId    string

	ConfigurationId int
}

// Find utility method to look for a specific project.
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*PolicyBody, error) {
	policies, err := List(ctx, cli, ListOptions{
		Organization: opts.Organization,
		ProjectId:    opts.ProjectId,
	})
	if err != nil {
		return nil, err
	}

	for _, policy := range policies.Value {
		if policy.ID == opts.ConfigurationId {
			return policy, nil
		}
	}

	return nil, &httplib.StatusError{
		StatusCode: http.StatusNotFound,
		Inner:      fmt.Errorf("policy configuration with ID %d not found in project %s/%s", opts.ConfigurationId, opts.Organization, opts.ProjectId),
	}
}
