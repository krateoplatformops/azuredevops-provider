package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Selectors selects an object.
type Selector struct {
	// Name is the name of a connector config.
	Name string `json:"name"`
	// Namespace is the namespace where the connector config belongs.
	Namespace string `json:"namespace"`
}

// Pipeline defines the desired state of Pipeline
type PipelineSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *Selector `json:"connectorConfigRef,omitempty"`

	// Organization: the organization name.
	// +immutable
	Organization string `json:"organization"`

	// Name: the name of the pipeline.
	// +immutable
	Name string `json:"name"`

	// Folder: the pipeline folder.
	Folder string `json:"folder,omitempty"`

	// ConfigurationType: Type of configuration.
	// +optional
	ConfigurationType *string `json:"configurationType,omitempty"`

	//DefinitionPath: The folder path of the definition.
	DefinitionPath *string `json:"definitionPath,omitempty"`

	// Project: TeamProject name or ID.
	// +optional
	Project *string `json:"project,omitempty"`

	// PojectRef - A reference to a TeamProject.
	// +optional
	PojectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// RepositoryRef: reference to the repository.
	RepositoryRef *Selector `json:"repositoryRef,omitempty"`

	// RepositoryType: Type of repository (default: azureReposGit).
	// +optional
	RepositoryType *string `json:"repositoryType,omitempty"`
}

type PipelineStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	// Pipeline ID
	Id *string `json:"id,omitempty"`
	// Revision number
	Revision *int `json:"revision,omitempty"`
	// URL of the pipeline
	Url *string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}
