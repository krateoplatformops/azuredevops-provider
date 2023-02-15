package projects

import (
	"gihtub.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops"
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

// A named value associated with a project.
type ProjectProperty struct {
	// The name of the property.
	Name *string `json:"name,omitempty"`
	// The value of the property.
	Value any `json:"value,omitempty"`
}

type ProjectVisibility string

const (
	VisibilityPrivate ProjectVisibility = "private"
	VisibilityPublic  ProjectVisibility = "public"
)

// Contains information describing a project.
type ProjectInfo struct {
	// The abbreviated name of the project.
	Abbreviation *string `json:"abbreviation,omitempty"`
	// The description of the project.
	Description *string `json:"description,omitempty"`
	// The id of the project.
	Id *string `json:"id,omitempty"`
	// The time that this project was last updated.
	LastUpdateTime *azuredevops.Time `json:"lastUpdateTime,omitempty"`
	// The name of the project.
	Name *string `json:"name,omitempty"`
	// A set of name-value pairs storing additional property data related to the project.
	Properties *[]ProjectProperty `json:"properties,omitempty"`
	// The current revision of the project.
	Revision *uint64 `json:"revision,omitempty"`
	// The current state of the project.
	State *ProjectState `json:"state,omitempty"`
	// A Uri that can be used to refer to this project.
	Uri *string `json:"uri,omitempty"`
	// The version number of the project.
	Version *uint64 `json:"version,omitempty"`
	// Indicates whom the project is visible to.
	Visibility *ProjectVisibility `json:"visibility,omitempty"`
}

type WebApiTeamRef struct {
	// Team (Identity) Guid. A Team Foundation ID.
	Id *string `json:"id,omitempty"`
	// Team name
	Name *string `json:"name,omitempty"`
	// Team REST API Url
	Url *string `json:"url,omitempty"`
}

// Represents a Team Project object.
type TeamProject struct {
	// Project abbreviation.
	Abbreviation *string `json:"abbreviation,omitempty"`
	// Url to default team identity image.
	DefaultTeamImageUrl *string `json:"defaultTeamImageUrl,omitempty"`
	// The project's description (if any).
	Description *string `json:"description,omitempty"`
	// Project identifier.
	Id *string `json:"id,omitempty"`
	// Project last update time.
	LastUpdateTime *azuredevops.Time `json:"lastUpdateTime,omitempty"`
	// Project name.
	Name *string `json:"name,omitempty"`
	// Project revision.
	Revision *uint64 `json:"revision,omitempty"`
	// Project state.
	State *ProjectState `json:"state,omitempty"`
	// Url to the full version of the object.
	Url *string `json:"url,omitempty"`
	// Project visibility.
	Visibility *ProjectVisibility `json:"visibility,omitempty"`
	// The links to other objects related to this object.
	Links any `json:"_links,omitempty"`
	// Set of capabilities this project has (such as process template & version control).
	Capabilities *map[string]map[string]string `json:"capabilities,omitempty"`
	// The shallow ref to the default team.
	DefaultTeam *WebApiTeamRef `json:"defaultTeam,omitempty"`
}

// Reference for an async operation.
type OperationReference struct {
	// Unique identifier for the operation.
	Id *string `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *string `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
}

// The status of an operation.
type OperationStatus string

const (
	StatusNotSet     OperationStatus = "notSet"
	StatusQueued     OperationStatus = "queued"
	StatusInProgress OperationStatus = "inProgress"
	StatusCancelled  OperationStatus = "cancelled"
	StatusSucceeded  OperationStatus = "succeded"
	StatusFailed     OperationStatus = "failed"
)
