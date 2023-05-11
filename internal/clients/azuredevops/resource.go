package azuredevops

type Resource struct {
	// Id of the resource.
	Id *string `json:"id,omitempty"`
	// Name of the resource.
	Name *string `json:"name,omitempty"`
	// Type of the resource.
	Type *string `json:"type,omitempty"`
}
