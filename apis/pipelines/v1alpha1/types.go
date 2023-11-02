package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Pipeline defines the desired state of Pipeline
type PipelineSpec struct {
	rtv1.ManagedSpec `json:",inline"`

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

	// RepositoryRef: reference to the repository.
	RepositoryRef *rtv1.Reference `json:"repositoryRef,omitempty"`

	// RepositoryType: Type of repository (default: azureReposGit).
	// +kubebuilder:default=azureReposGit
	// +optional
	RepositoryType *string `json:"repositoryType,omitempty"`

	// ProjectRef - A reference to a TeamProject.
	// +optional
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
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
