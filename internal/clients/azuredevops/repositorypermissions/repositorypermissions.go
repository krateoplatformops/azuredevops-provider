package repositoryspermissions

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
	"github.com/lucasepe/httplib"
)

const securityNamespaceId = "2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87"

type IdentityPermission struct {
	Descriptor string `json:"descriptor"`
	Allow      int    `json:"allow"`
	Deny       int    `json:"deny"`
}

type PermissionResponse struct {
	Count int                  `json:"count"`
	Value []IdentityPermission `json:"value"`
}

type AccessControlEntry struct {
	Descriptor string `json:"descriptor"`
	Allow      int    `json:"allow"`
	Deny       int    `json:"deny"`
}
type AccessControlUpdate struct {
	Merge                bool                 `json:"merge"`
	Token                string               `json:"token"`
	AccessControlEntries []AccessControlEntry `json:"accessControlEntries"`
}

type PermissionBit int

const (
	Administer PermissionBit = 1 << iota
	GenericRead
	GenericContribute
	ForcePush
	CreateBranch
	CreateTag
	ManageNote
	PolicyExempt
	CreateRepository
	DeleteRepository
	RenameRepository
	EditPolicies
	RemoveOthersLocks
	ManagePermissions
	PullRequestContribute
	PullRequestBypassPolicy
	ViewAdvSecAlerts
	DismissAdvSecAlerts
	ManageAdvSecScanning
)

func (p PermissionBit) String() string {
	switch p {
	case Administer:
		return "administerpermission"
	case GenericRead:
		return "genericread"
	case GenericContribute:
		return "genericcontribute"
	case ForcePush:
		return "forcepush"
	case CreateBranch:
		return "createbranch"
	case CreateTag:
		return "createtag"
	case ManageNote:
		return "managenote"
	case PolicyExempt:
		return "policyexempt"
	case CreateRepository:
		return "createrepository"
	case DeleteRepository:
		return "deleterepository"
	case RenameRepository:
		return "renamerepository"
	case EditPolicies:
		return "editpolicies"
	case RemoveOthersLocks:
		return "removeotherslocks"
	case ManagePermissions:
		return "managepermissions"
	case PullRequestContribute:
		return "pullrequestcontribute"
	case PullRequestBypassPolicy:
		return "pullrequestbypasspolicy"
	case ViewAdvSecAlerts:
		return "viewadvsecalerts"
	case DismissAdvSecAlerts:
		return "dismissadvsecalerts"
	case ManageAdvSecScanning:
		return "manageadvsecscanning"
	default:
		return ""
	}
}

// Value returns the int value of the permission bit specified by perm string or -1 if not found
func PermissionBitValue(perm string) int {
	perm = strings.ToLower(perm)
	switch perm {
	case "administerpermission":
		return int(Administer)
	case "genericread":
		return int(GenericRead)
	case "genericcontribute":
		return int(GenericContribute)
	case "forcepush":
		return int(ForcePush)
	case "createbranch":
		return int(CreateBranch)
	case "createtag":
		return int(CreateTag)
	case "managenote":
		return int(ManageNote)
	case "policyexempt":
		return int(PolicyExempt)
	case "createrepository":
		return int(CreateRepository)
	case "deleterepository":
		return int(DeleteRepository)
	case "renamerepository":
		return int(RenameRepository)
	case "editpolicies":
		return int(EditPolicies)
	case "removeotherslocks":
		return int(RemoveOthersLocks)
	case "managepermissions":
		return int(ManagePermissions)
	case "pullrequestcontribute":
		return int(PullRequestContribute)
	case "pullrequestbypasspolicy":
		return int(PullRequestBypassPolicy)
	case "viewadvsecalerts":
		return int(ViewAdvSecAlerts)
	case "dismissadvsecalerts":
		return int(DismissAdvSecAlerts)
	case "manageadvsecscanning":
		return int(ManageAdvSecScanning)
	default:
		return -1
	}
}

func CreateToken(projectId, repoId string) string {
	return path.Join("repoV2/", projectId, repoId)
}

type GetOptions struct {
	Organization string `json:"organization"`
	Descriptor   string `json:"descriptor"`
	Token        string `json:"token"`
}

func Get(ctx context.Context, cli *azuredevops.Client, opts GetOptions) (*PermissionResponse, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/accesscontrolentries", securityNamespaceId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	getOptions := &AccessControlUpdate{
		Merge: true,
		Token: opts.Token,
		AccessControlEntries: []AccessControlEntry{
			{
				Descriptor: opts.Descriptor,
				Allow:      0,
				Deny:       0,
			},
		},
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(getOptions))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &PermissionResponse{
		Count: 0,
		Value: []IdentityPermission{},
	}
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

type UpdateOptions struct {
	Organization          string               `json:"organization"`
	ResourceAuthorization *AccessControlUpdate `json:"resourceAuthorization"`
}

// Authorizes/Unauthorizes a list of definitions for a given resource.
// POST https://dev.azure.com/{organization}/_apis/accesscontrolentries/{securityNamespaceId}?api-version=7.0
func Update(ctx context.Context, cli *azuredevops.Client, opts UpdateOptions) (*PermissionResponse, error) {
	ubo := httplib.URLBuilderOptions{
		BaseURL: cli.BaseURL(azuredevops.Default),
		Path:    path.Join(opts.Organization, "_apis/accesscontrolentries", securityNamespaceId),
		Params:  []string{azuredevops.ApiVersionKey, azuredevops.ApiVersionVal},
	}

	uri, err := httplib.NewURLBuilder(ubo).Build()
	if err != nil {
		return nil, err
	}

	req, err := httplib.Post(uri.String(), httplib.ToJSON(opts.ResourceAuthorization))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req = req.WithContext(ctx)

	apiErr := &azuredevops.APIError{}
	val := &PermissionResponse{
		Count: 0,
		Value: []IdentityPermission{},
	}
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
