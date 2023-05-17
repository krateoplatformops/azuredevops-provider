package pools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/lucasepe/httplib"
)

// An organization-level grouping of agents.
type TaskAgentPool struct {
	Id *int `json:"id,omitempty"`
	// Gets or sets a value indicating whether or not this pool is managed by the service.
	IsHosted *bool `json:"isHosted,omitempty"`
	// Determines whether the pool is legacy.
	IsLegacy *bool   `json:"isLegacy,omitempty"`
	Name     *string `json:"name,omitempty"`
	// Additional pool settings and details
	// [none, elasticPool, singleUseAgents, preserveAgentOnJobFailure]
	Options *string `json:"options,omitempty"`
	// Gets or sets the type of the pool
	PoolType *string `json:"poolType,omitempty"`
	Scope    *string `json:"scope,omitempty"`
	// Gets the current size of the pool.
	Size int `json:"size,omitempty"`
	// The ID of the associated agent cloud.
	AgentCloudId *int `json:"agentCloudId,omitempty"`
	// Whether or not a queue should be automatically provisioned for each project collection.
	AutoProvision *bool `json:"autoProvision,omitempty"`
	// Whether or not the pool should autosize itself based on the Agent Cloud Provider settings.
	AutoSize *bool `json:"autoSize,omitempty"`
	// Whether or not agents in this pool are allowed to automatically update
	AutoUpdate *bool `json:"autoUpdate,omitempty"`
	// Creator of the pool. The creator of the pool is automatically added into the administrators group for the pool on creation.
	CreatedBy *azuredevops.IdentityRef `json:"createdBy,omitempty"`
	// The date/time of the pool creation.
	CreatedOn *azuredevops.Time `json:"createdOn,omitempty"`
	// Owner or administrator of the pool.
	Owner      *azuredevops.IdentityRef `json:"owner,omitempty"`
	Properties interface{}              `json:"properties,omitempty"`
	// Target parallelism - Only applies to agent pools that are backed by pool providers. It will be null for regular pools.
	TargetSize *int `json:"targetSize,omitempty"`
}

type FindOptions struct {
	// (required) Name of the organization
	Organization string
	// (optional) Filter by name
	PoolName string
	// (optional) Filter by agent pool properties (comma-separated)
	Properties []string
	// (optional) Filter by pool type ["automation", "deployment"]
	PoolType *string
	// (optional) Filter by whether the calling user has use or manage permissions
	//ActionFilter *string
}

type FindResult struct {
	Count  int             `json:"count"`
	Values []TaskAgentPool `json:"value,omitempty"`
}

// GET https://dev.azure.com/{organization}/_apis/distributedtask/pools?poolName={poolName}&properties={properties}&poolType={poolType}&actionFilter={actionFilter}&api-version=7.0
func Find(ctx context.Context, cli *azuredevops.Client, opts FindOptions) ([]TaskAgentPool, error) {
	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	if len(opts.PoolName) > 0 {
		params = append(params, "poolName", opts.PoolName)
	}
	if len(opts.Properties) > 0 {
		params = append(params, "properties", strings.Join(opts.Properties, ","))
	}
	if opts.PoolType != nil {
		params = append(params, "poolType", helpers.String(opts.PoolType))
	}

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
			Path:    path.Join(opts.Organization, "_apis/distributedtask/pools"),
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
	val := []TaskAgentPool{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		AuthMethod: cli.AuthMethod(),
		Verbose:    cli.Verbose(),
		ResponseHandler: func(res *http.Response) error {
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			all := &FindResult{}
			if err = json.Unmarshal(data, &all); err != nil {
				return err
			}

			val = append(val, all.Values...)

			return nil
		},

		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})

	return val, err
}
