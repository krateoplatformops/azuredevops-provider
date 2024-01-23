package v1alpha2

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceType string

const (
	GitRepository ResourceType = "repository"
	Environment   ResourceType = "environment"
	Queue         ResourceType = "queue"
	TeamProject   ResourceType = "teamproject"
	Endpoint      ResourceType = "endpoint"
)

type Resource struct {
	// Type of the resource.
	Type *string `json:"type,omitempty"`

	// ResourceRef - Reference to the resource to authorize.
	ResourceRef *rtv1.Reference `json:"resourceRef,omitempty"`
}
type PipelineAuthorization struct {
	// Authorized - Whether or not this pipeline is authorized for use.
	Authorized bool `json:"authorized"`
	// PipelineRef - Reference to pipeline to authorize/unauthorize.
	PipelineRef *rtv1.Reference `json:"pipelineRef"`
}

// PipelinePermission defines the desired state of PipelinePermission
type PipelinePermissionSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ProjectRef - Reference to the project to authorize.
	// +required
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// Resource - Resource to authorize.
	// +required
	Resource *Resource `json:"resource,omitempty"`

	// Pipelines - List of pipeline names to authorize.
	// +optional
	Pipelines []PipelineAuthorization `json:"pipelines,omitempty"`

	// AuthorizeAll - If true, authorize all pipelines in the project.
	// +omitempty
	AuthorizeAll *bool `json:"authorizeAll,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
}

type PipelinePermissionStatus struct {
	rtv1.ManagedStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// PipelinePermission is the Schema for the PipelinePermissions API
type PipelinePermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelinePermissionSpec   `json:"spec,omitempty"`
	Status PipelinePermissionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelinePermissionList contains a list of PipelinePermission
type PipelinePermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelinePermission `json:"items"`
}
