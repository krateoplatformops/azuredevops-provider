package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
)

type ProjectReference struct {
	// Gets or sets id of the project.
	Id string `json:"id,omitempty"`
	// Gets or sets name of the project.
	Name string `json:"name,omitempty"`
	// Gets or sets visibility of the project.
	Visibility string `json:"visibility,omitempty"`
}

type UpstreamStatusDetail struct {
	// Provides a human-readable reason for the status of the upstream.
	Reason *string `json:"reason,omitempty"`
}

// Upstream source definition, including its Identity, package type, and other associated information.
type UpstreamSource struct {
	// UTC date that this upstream was deleted.
	DeletedDate *azuredevops.Time `json:"deletedDate,omitempty"`
	// Locator for connecting to the upstream source in a user friendly format, that may potentially change over time
	DisplayLocation *string `json:"displayLocation,omitempty"`
	// Identity of the upstream source.
	Id *string `json:"id,omitempty"`
	// For an internal upstream type, track the Azure DevOps organization that contains it.
	InternalUpstreamCollectionId *string `json:"internalUpstreamCollectionId,omitempty"`
	// For an internal upstream type, track the feed id being referenced.
	InternalUpstreamFeedId *string `json:"internalUpstreamFeedId,omitempty"`
	// For an internal upstream type, track the project of the feed being referenced.
	InternalUpstreamProjectId *string `json:"internalUpstreamProjectId,omitempty"`
	// For an internal upstream type, track the view of the feed being referenced.
	InternalUpstreamViewId *string `json:"internalUpstreamViewId,omitempty"`
	// Consistent locator for connecting to the upstream source.
	Location *string `json:"location,omitempty"`
	// Display name.
	Name *string `json:"name,omitempty"`
	// Package type associated with the upstream source.
	Protocol *string `json:"protocol,omitempty"`
	// The identity of the service endpoint that holds credentials to use when accessing the upstream.
	ServiceEndpointId *string `json:"serviceEndpointId,omitempty"`
	// Specifies the projectId of the Service Endpoint.
	ServiceEndpointProjectId *string `json:"serviceEndpointProjectId,omitempty"`
	// Specifies the status of the upstream.
	// [ok, disabled]
	Status *string `json:"status,omitempty"`
	// Provides a human-readable reason for the status of the upstream.
	StatusDetails []UpstreamStatusDetail `json:"statusDetails,omitempty"`
	// Source type, such as Public or Internal.
	// [public, internal]
	UpstreamSourceType *string `json:"upstreamSourceType,omitempty"`
}

// Permissions for a feed.
type FeedPermission struct {
	// Display name for the identity.
	DisplayName *string `json:"displayName,omitempty"`
	// Identity associated with this role.
	IdentityDescriptor *string `json:"identityDescriptor,omitempty"`
	// Id of the identity associated with this role.
	IdentityId *string `json:"identityId,omitempty"`
	// Boolean indicating whether the role is inherited or set directly.
	IsInheritedRole *bool `json:"isInheritedRole,omitempty"`
	// The role for this identity on a feed.
	// [custom, none, reader, contributor, administrator, collaborator]
	Role *string `json:"role,omitempty"`
}

// A view on top of a feed.
type FeedView struct {
	// Related REST links.
	Links interface{} `json:"_links,omitempty"`
	// Id of the view.
	Id *string `json:"id,omitempty"`
	// Name of the view.
	Name *string `json:"name,omitempty"`
	// Type of view.
	// [none, release, implicit]
	Type *string `json:"type,omitempty"`
	// Url of the view.
	Url *string `json:"url,omitempty"`
	// Visibility status of the view.
	// [private, collection, organization, aadTenant]
	Visibility *string `json:"visibility,omitempty"`
}

// A container for artifacts.
type Feed struct {
	// Supported capabilities of a feed.
	// [none, upstreamV2, underMaintenance, defaultCapabilities]
	Capabilities *string `json:"capabilities,omitempty"`
	// This will either be the feed GUID or the feed GUID and view GUID depending on how the feed was accessed.
	FullyQualifiedId *string `json:"fullyQualifiedId,omitempty"`
	// Full name of the view, in feed@view format.
	FullyQualifiedName *string `json:"fullyQualifiedName,omitempty"`
	// A GUID that uniquely identifies this feed.
	Id *string `json:"id,omitempty"`
	// If set, all packages in the feed are immutable.  It is important to note that feed views are immutable; therefore, this flag will always be set for views.
	IsReadOnly bool `json:"isReadOnly,omitempty"`
	// A name for the feed. feed names must follow these rules:
	// - Must not exceed 64 characters
	// - Must not contain whitespaces
	// - Must not start with an underscore or a period
	// - Must not end with a period
	// - Must not contain any of the following illegal characters: <![CDATA[ @, ~, ;, {, }, \, +, =, <, >, |, /, \\, ?, :, &, $, *, \", #, [, ] ]]>
	Name string `json:"name,omitempty"`
	// The project that this feed is associated with.
	Project *ProjectReference `json:"project,omitempty"`
	// This should always be true. Setting to false will override all sources in UpstreamSources.
	UpstreamEnabled bool `json:"upstreamEnabled,omitempty"`
	// A list of sources that this feed will fetch packages from.  An empty list indicates that this feed will not search any additional sources for packages.
	UpstreamSources []UpstreamSource `json:"upstreamSources,omitempty"`
	// Definition of the view.
	View *FeedView `json:"view,omitempty"`
	// View Id.
	ViewId *string `json:"viewId,omitempty"`
	// View name.
	ViewName *string `json:"viewName,omitempty"`
	// Related REST links.
	Links interface{} `json:"_links,omitempty"`
	// If set, this feed supports generation of package badges.
	BadgesEnabled bool `json:"badgesEnabled,omitempty"`
	// The view that the feed administrator has indicated is the default experience for readers.
	DefaultViewId *string `json:"defaultViewId,omitempty"`
	// The date that this feed was deleted.
	DeletedDate *azuredevops.Time `json:"deletedDate,omitempty"`
	// A description for the feed.  Descriptions must not exceed 255 characters.
	Description *string `json:"description,omitempty"`
	// If set, the feed will hide all deleted/unpublished versions
	HideDeletedPackageVersions bool `json:"hideDeletedPackageVersions,omitempty"`
	// The date that this feed was permanently deleted.
	PermanentDeletedDate *azuredevops.Time `json:"permanentDeletedDate,omitempty"`
	// Explicit permissions for the feed.
	Permissions []FeedPermission `json:"permissions,omitempty"`
	// The date that this feed is scheduled to be permanently deleted.
	ScheduledPermanentDeleteDate *azuredevops.Time `json:"scheduledPermanentDeleteDate,omitempty"`
	// If set, time that the UpstreamEnabled property was changed. Will be null if UpstreamEnabled was never changed after Feed creation.
	UpstreamEnabledChangedDate *azuredevops.Time `json:"upstreamEnabledChangedDate,omitempty"`
	// The URL of the base feed in GUID form.
	Url *string `json:"url,omitempty"`
}

// Update a feed definition with these new values.
type FeedUpdate struct {
	// If set, the feed will allow upload of packages that exist on the upstream
	AllowUpstreamNameConflict bool `json:"allowUpstreamNameConflict,omitempty"`
	// If set, this feed supports generation of package badges.
	BadgesEnabled bool `json:"badgesEnabled,omitempty"`
	// The view that the feed administrator has indicated is the default experience for readers.
	DefaultViewId *string `json:"defaultViewId,omitempty"`
	// A description for the feed.  Descriptions must not exceed 255 characters.
	Description *string `json:"description,omitempty"`
	// If set, feed will hide all deleted/unpublished versions
	HideDeletedPackageVersions bool `json:"hideDeletedPackageVersions,omitempty"`
	// A GUID that uniquely identifies this feed.
	Id *string `json:"id,omitempty"`
	// A name for the feed.
	Name string `json:"name,omitempty"`
	// If set, the feed can proxy packages from an upstream feed
	UpstreamEnabled bool `json:"upstreamEnabled,omitempty"`
	// A list of sources that this feed will fetch packages from.
	// An empty list indicates that this feed will not search any additional sources for packages.
	UpstreamSources []UpstreamSource `json:"upstreamSources,omitempty"`
}

// Options for the Get feed function
type GetOptions struct {
	// (required) Name or Id of the feed.
	FeedId string
	// (required) Name of the organization
	Organization string
	// (optional) Project ID or project name
	Project string
	// (optional) Include upstreams that have been deleted in the response.
	//IncludeDeletedUpstreams *bool
}

// Get the settings for a specific feed.
// GET https://feeds.dev.azure.com/{organization}/{project}/_apis/packaging/feeds/{feedId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*Feed, error) {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/packaging/feeds/", opts.FeedId)
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/packaging/feeds/", opts.FeedId)
	}

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Feeds),
		Path:    fullPath,
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
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
	val := &Feed{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	if val != nil && reflect.DeepEqual(*val, Feed{}) {
		return nil, err
	}

	return val, err
}

// Options for the List function
type ListOptions struct {
	// (required) Name of the organization
	Organization string
	// (optional) Project ID or project name
	Project string
	// (optional) Filter by this role, either Administrator(4), Contributor(3), or Reader(2) level permissions.
	FeedRole string
	// (optional) Resolve names if true
	IncludeUrls bool
}

type ListResult struct {
	Count int    `json:"count"`
	Feeds []Feed `json:"value,omitempty"`
}

// Get all feeds in an account where you have the provided role access.
// GET https://feeds.dev.azure.com/{organization}/{project}/_apis/packaging/feeds?api-version=7.0
func List(ctx context.Context, cli *azuredevops.Client, opts ListOptions) ([]Feed, error) {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/packaging/feeds")
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/packaging/feeds")
	}

	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	if len(opts.FeedRole) > 0 {
		params = append(params, "feedRole", opts.FeedRole)
	}
	params = append(params, "includeUrls", fmt.Sprintf("%t", opts.IncludeUrls))

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Feeds),
			Path:    fullPath,
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
	val := []Feed{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		ResponseHandler: func(res *http.Response) error {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			all := &ListResult{}
			if err = json.Unmarshal(data, &all); err != nil {
				return err
			}

			val = append(val, all.Feeds...)

			return nil
		},

		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}

type FindOptions struct {
	// Name of the organization
	Organization string
	Project      string
	FeedName     string
}

func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) (*Feed, error) {
	all, err := List(ctx, cli, ListOptions{
		Organization: opts.Organization,
		Project:      opts.Project,
		IncludeUrls:  true,
	})
	if err != nil {
		return nil, err
	}

	for _, el := range all {
		if el.Name == opts.FeedName {
			return &el, nil
		}
	}

	return nil, nil
}

// Options for the Create feed function
type CreateOptions struct {
	// Name of the organization
	Organization string
	// (required) A JSON object containing both required and optional attributes for the feed.
	// Name is the only required value.
	Feed *Feed
	// (optional) Project ID or project name
	Project string
}

// Create a feed, a container for various package types.
// POST https://feeds.dev.azure.com/{organization}/{project}/_apis/packaging/feeds?api-version=7.0
func Create(ctx context.Context, cli *azuredevops.Client, opts CreateOptions) (*Feed, error) {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/packaging/feeds")
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/packaging/feeds")
	}

	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Feeds),
		Path:    fullPath,
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.Feed))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &Feed{}
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

type UpdateOptions struct {
	Organization string
	Project      string
	// (required) A JSON object containing the feed settings to be updated.
	FeedUpdate *FeedUpdate
	// (required) Name or Id of the feed.
	FeedId string
}

// Update an existing project's name, abbreviation, description, or restore a project.
// PATCH https://feeds.dev.azure.com/{organization}/{project}/_apis/packaging/feeds/{feedId}?api-version=7.0
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*Feed, error) {
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Feeds),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/packaging/feeds", opts.FeedId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Patch(uri.String(), httplib.ToJSON(opts.FeedUpdate))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &Feed{}
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
	Project      string
	FeedId       string
}

// Delete a feed.
// DELETE https://feeds.dev.azure.com/{organization}/{project}/_apis/packaging/feeds/{feedId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/packaging/feeds/", opts.FeedId)
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/packaging/feeds/", opts.FeedId)
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Feeds),
		Path:    fullPath,
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
