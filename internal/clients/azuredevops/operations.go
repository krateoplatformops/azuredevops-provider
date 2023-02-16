package azuredevops

type OperationResultReference struct {
	// URL to the operation result.
	ResultUrl *string `json:"resultUrl,omitempty"`
}

// Contains information about the progress or result of an async operation.
type Operation struct {
	// Unique identifier for the operation.
	Id *string `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *string `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Links to other related objects.
	Links interface{} `json:"_links,omitempty"`
	// Detailed messaged about the status of an operation.
	DetailedMessage *string `json:"detailedMessage,omitempty"`
	// Result message for an operation.
	ResultMessage *string `json:"resultMessage,omitempty"`
	// URL to the operation result.
	ResultUrl *OperationResultReference `json:"resultUrl,omitempty"`
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
