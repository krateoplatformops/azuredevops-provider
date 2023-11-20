package queues

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
	"github.com/lucasepe/httplib"
)

type TaskAgentPoolReference struct {
	Id *int `json:"id,omitempty"`
	// Gets or sets a value indicating whether or not this pool is managed by the service.
	IsHosted *bool `json:"isHosted,omitempty"`
	// Determines whether the pool is legacy.
	IsLegacy *bool  `json:"isLegacy,omitempty"`
	Name     string `json:"name,omitempty"`
	// Additional pool settings and details
	// [none, elasticPool, singleUseAgents, preserveAgentOnJobFailure]
	Options *string `json:"options,omitempty"`
	// Gets or sets the type of the pool
	// [automation, deployment]
	PoolType *string `json:"poolType,omitempty"`
	Scope    string  `json:"scope,omitempty"`
	// Gets the current size of the pool.
	Size *int `json:"size,omitempty"`
}

// A reference to an agent.
type TaskAgentReference struct {
	Links interface{} `json:"_links,omitempty"`
	// This agent's access point.
	AccessPoint *string `json:"accessPoint,omitempty"`
	// Whether or not this agent should run jobs.
	Enabled *bool `json:"enabled,omitempty"`
	// Identifier of the agent.
	Id *int `json:"id,omitempty"`
	// Name of the agent.
	Name *string `json:"name,omitempty"`
	// Agent OS.
	OsDescription *string `json:"osDescription,omitempty"`
	// Provisioning state of this agent.
	ProvisioningState *string `json:"provisioningState,omitempty"`
	// Whether or not the agent is online.
	Status string `json:"status,omitempty"`
	// Agent version.
	Version *string `json:"version,omitempty"`
}

type TaskAgentCloudRequest struct {
	Agent                *TaskAgentReference     `json:"agent,omitempty"`
	AgentCloudId         *int                    `json:"agentCloudId,omitempty"`
	AgentConnectedTime   *azuredevops.Time       `json:"agentConnectedTime,omitempty"`
	AgentData            interface{}             `json:"agentData,omitempty"`
	AgentSpecification   interface{}             `json:"agentSpecification,omitempty"`
	Pool                 *TaskAgentPoolReference `json:"pool,omitempty"`
	ProvisionedTime      *azuredevops.Time       `json:"provisionedTime,omitempty"`
	ProvisionRequestTime *azuredevops.Time       `json:"provisionRequestTime,omitempty"`
	ReleaseRequestTime   *azuredevops.Time       `json:"releaseRequestTime,omitempty"`
	RequestId            string                  `json:"requestId,omitempty"`
}

type TaskOrchestrationOwner struct {
	Links interface{} `json:"_links,omitempty"`
	Id    *int        `json:"id,omitempty"`
	Name  *string     `json:"name,omitempty"`
}

// A job request for an agent.
type TaskAgentJobRequest struct {
	AgentSpecification interface{} `json:"agentSpecification,omitempty"`
	// The date/time this request was assigned.
	AssignTime *azuredevops.Time `json:"assignTime,omitempty"`
	// Additional data about the request.
	Data *map[string]string `json:"data,omitempty"`
	// The pipeline definition associated with this request
	Definition *TaskOrchestrationOwner `json:"definition,omitempty"`
	// A list of demands required to fulfill this request.
	Demands *[]interface{} `json:"demands,omitempty"`
	// The date/time this request was finished.
	FinishTime *azuredevops.Time `json:"finishTime,omitempty"`
	// The host which triggered this request.
	HostId *string `json:"hostId,omitempty"`
	// ID of the job resulting from this request.
	JobId *string `json:"jobId,omitempty"`
	// Name of the job resulting from this request.
	JobName *string `json:"jobName,omitempty"`
	// The deadline for the agent to renew the lock.
	LockedUntil            *azuredevops.Time    `json:"lockedUntil,omitempty"`
	MatchedAgents          []TaskAgentReference `json:"matchedAgents,omitempty"`
	MatchesAllAgentsInPool *bool                `json:"matchesAllAgentsInPool,omitempty"`
	OrchestrationId        *string              `json:"orchestrationId,omitempty"`
	// The pipeline associated with this request
	Owner     *TaskOrchestrationOwner `json:"owner,omitempty"`
	PlanGroup *string                 `json:"planGroup,omitempty"`
	// Internal ID for the orchestration plan connected with this request.
	PlanId *string `json:"planId,omitempty"`
	// Internal detail representing the type of orchestration plan.
	PlanType *string `json:"planType,omitempty"`
	// The ID of the pool this request targets
	PoolId   *int `json:"poolId,omitempty"`
	Priority *int `json:"priority,omitempty"`
	// The ID of the queue this request targets
	QueueId *int `json:"queueId,omitempty"`
	// The date/time this request was queued.
	QueueTime *azuredevops.Time `json:"queueTime,omitempty"`
	// The date/time this request was receieved by an agent.
	ReceiveTime *azuredevops.Time `json:"receiveTime,omitempty"`
	// ID of the request.
	RequestId *uint64 `json:"requestId,omitempty"`
	// The agent allocated for this request.
	ReservedAgent *TaskAgentReference `json:"reservedAgent,omitempty"`
	// The result of this request.
	// [succeeded, succeededWithIssues, failed, canceled, skipped, abandoned]
	Result *string `json:"result,omitempty"`
	// Scope of the pipeline; matches the project ID.
	ScopeId *string `json:"scopeId,omitempty"`
	// The service which owns this request.
	ServiceOwner  *string `json:"serviceOwner,omitempty"`
	StatusMessage *string `json:"statusMessage,omitempty"`
	UserDelayed   *bool   `json:"userDelayed,omitempty"`
}

// Represents the public key portion of an RSA asymmetric key.
type TaskAgentPublicKey struct {
	// Gets or sets the exponent for the public key.
	Exponent *[]byte `json:"exponent,omitempty"`
	// Gets or sets the modulus for the public key.
	Modulus *[]byte `json:"modulus,omitempty"`
}

// Provides data necessary for authorizing the agent using OAuth 2.0 authentication flows.
type TaskAgentAuthorization struct {
	// Endpoint used to obtain access tokens from the configured token service.
	AuthorizationUrl *string `json:"authorizationUrl,omitempty"`
	// Client identifier for this agent.
	ClientId *string `json:"clientId,omitempty"`
	// Public key used to verify the identity of this agent.
	PublicKey *TaskAgentPublicKey `json:"publicKey,omitempty"`
}

type TaskAgentUpdateReason struct {
	// [manual, minAgentVersionRequired, downgrade]
	Code *string `json:"code,omitempty"`
}

type PackageVersion struct {
	Major *int `json:"major,omitempty"`
	Minor *int `json:"minor,omitempty"`
	Patch *int `json:"patch,omitempty"`
}

// Details about an agent update.
type TaskAgentUpdate struct {
	// Current state of this agent update.
	CurrentState *string `json:"currentState,omitempty"`
	// Reason for this update.
	Reason *TaskAgentUpdateReason `json:"reason,omitempty"`
	// Identity which requested this update.
	RequestedBy *azuredevops.IdentityRef `json:"requestedBy,omitempty"`
	// Date on which this update was requested.
	RequestTime *azuredevops.Time `json:"requestTime,omitempty"`
	// Source agent version of the update.
	SourceVersion *PackageVersion `json:"sourceVersion,omitempty"`
	// Target agent version of the update.
	TargetVersion *PackageVersion `json:"targetVersion,omitempty"`
}

// A task agent.
type TaskAgent struct {
	Links interface{} `json:"_links,omitempty"`
	// This agent's access point.
	AccessPoint *string `json:"accessPoint,omitempty"`
	// Whether or not this agent should run jobs.
	Enabled *bool `json:"enabled,omitempty"`
	// Identifier of the agent.
	Id *int `json:"id,omitempty"`
	// Name of the agent.
	Name *string `json:"name,omitempty"`
	// Agent OS.
	OsDescription *string `json:"osDescription,omitempty"`
	// Provisioning state of this agent.
	ProvisioningState *string `json:"provisioningState,omitempty"`
	// Whether or not the agent is online.
	// [offline, online]
	Status string `json:"status,omitempty"`
	// Agent version.
	Version *string `json:"version,omitempty"`
	// The agent cloud request that's currently associated with this agent.
	AssignedAgentCloudRequest *TaskAgentCloudRequest `json:"assignedAgentCloudRequest,omitempty"`
	// The request which is currently assigned to this agent.
	AssignedRequest *TaskAgentJobRequest `json:"assignedRequest,omitempty"`
	// Authorization information for this agent.
	Authorization *TaskAgentAuthorization `json:"authorization,omitempty"`
	// Date on which this agent was created.
	CreatedOn *azuredevops.Time `json:"createdOn,omitempty"`
	// The last request which was completed by this agent.
	LastCompletedRequest *TaskAgentJobRequest `json:"lastCompletedRequest,omitempty"`
	// Maximum job parallelism allowed for this agent.
	MaxParallelism *int `json:"maxParallelism,omitempty"`
	// Pending update for this agent.
	PendingUpdate *TaskAgentUpdate `json:"pendingUpdate,omitempty"`
	Properties    interface{}      `json:"properties,omitempty"`
	// Date on which the last connectivity status change occurred.
	StatusChangedOn *azuredevops.Time `json:"statusChangedOn,omitempty"`
	// System-defined capabilities supported by this agent's host. Warning: To set capabilities use the PUT method, PUT will completely overwrite existing capabilities.
	SystemCapabilities map[string]string `json:"systemCapabilities,omitempty"`
	// User-defined capabilities supported by this agent's host. Warning: To set capabilities use the PUT method, PUT will completely overwrite existing capabilities.
	UserCapabilities map[string]string `json:"userCapabilities,omitempty"`
}

// An agent queue.
type TaskAgentQueue struct {
	// ID of the queue
	Id *int `json:"id,omitempty"`
	// Name of the queue
	Name string `json:"name,omitempty"`
	// Pool reference for this queue
	Pool *TaskAgentPoolReference `json:"pool,omitempty"`
	// Project ID
	ProjectId *string `json:"projectId,omitempty"`
}

type AddOptions struct {
	// (required) Name of the organization
	Organization string
	// (optional) Project ID or project name
	Project string
	// (required) Details about the queue to create
	Queue *TaskAgentQueue
	// (optional) Automatically authorize this queue when using YAML
	//AuthorizePipelines *bool
}

// POST https://dev.azure.com/{organization}/{project}/_apis/distributedtask/queues?api-version=7.0
func Add(ctx context.Context, cli *azuredevops.Client, opts AddOptions) (*TaskAgentQueue, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, opts.Project, "_apis/distributedtask/queues/"),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.Queue))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &TaskAgentQueue{}

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

type FindByNamesOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) A comma-separated list of agent names to retrieve
	QueueNames []string
	// (optional) Project ID or project name
	Project string
	// (optional) Filter by whether the calling user has use or manage permissions
	//ActionFilter *string
}

type FindResult struct {
	Count  int              `json:"count"`
	Values []TaskAgentQueue `json:"value,omitempty"`
}

// GET https://dev.azure.com/{organization}/{project}/_apis/distributedtask/queues?queueNames={queueNames}&api-version=7.0
func FindByNames(ctx context.Context, cli *azuredevops.Client, opts FindByNamesOptions) ([]TaskAgentQueue, error) {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/distributedtask/queues")
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/distributedtask/queues")
	}

	params := []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal}
	if len(opts.QueueNames) > 0 {
		params = append(params, "queueNames", strings.Join(opts.QueueNames, ","))
	}

	uri, err := httplib.NewURLBuilder(
		httplib.URLBuilderOptions{
			BaseURL: cli.BaseURL(azuredevops.Default),
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
	val := []TaskAgentQueue{}

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

	if len(val) == 0 {
		return nil, &httplib.StatusError{
			StatusCode: http.StatusNotFound,
			Inner:      fmt.Errorf("queue(s) [%s] not found", strings.Join(opts.QueueNames, ",")),
		}
	}

	return val, err
}

type GetOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) The agent queue to get information about
	QueueId int
	// (optional) Project ID or project name
	Project string
	// (optional) Filter by whether the calling user has use or manage permissions
	//ActionFilter *string
}

// GET https://dev.azure.com/{organization}/{project}/_apis/distributedtask/queues/{queueId}?api-version=7.0
func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*TaskAgentQueue, error) {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/distributedtask/queues/", fmt.Sprintf("%d", opts.QueueId))
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/distributedtask/queues/", fmt.Sprintf("%d", opts.QueueId))
	}

	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
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
	val := &TaskAgentQueue{}

	err = httplib.Fire(cli.HTTPClient(), req, httplib.FireOptions{
		Verbose:         cli.Verbose(),
		AuthMethod:      cli.AuthMethod(),
		ResponseHandler: httplib.FromJSON(val),
		Validators: []httplib.HandleResponseFunc{
			httplib.ErrorJSON(apiErr, http.StatusOK),
		},
	})
	if val != nil && reflect.DeepEqual(*val, TaskAgentQueue{}) {
		return nil, err
	}

	return val, err
}

type DeleteOptions struct {
	// (required) Name of the organization
	Organization string
	// (required) The agent queue to remove
	QueueId int
	// (optional) Project ID or project name
	Project string
}

// DELETE https://dev.azure.com/{organization}/{project}/_apis/distributedtask/queues/{queueId}?api-version=7.0
func Delete(ctx context.Context, cli *azuredevops.Client, opts DeleteOptions) error {
	var fullPath string
	if len(opts.Project) == 0 {
		fullPath = path.Join(opts.Organization, "_apis/distributedtask/queues/", fmt.Sprintf("%d", opts.QueueId))
	} else {
		fullPath = path.Join(opts.Organization, opts.Project, "_apis/distributedtask/queues/", fmt.Sprintf("%d", opts.QueueId))
	}
	uri, err := httplib.NewURLBuilder(httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
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
