package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	// Id of the resource.
	Id *string `json:"id,omitempty"`
	// Name of the resource.
	Name *string `json:"name,omitempty"`
	// Type of the resource.
	Type *string `json:"type,omitempty"`
}

// PipelinePermission defines the desired state of PipelinePermission
type PipelinePermissionSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// Project - TeamProject NAME OR ID.
	Project string `json:"project"`

	// Organization -  Organization NAME.
	Organization string `json:"organization"`

	Resource *Resource `json:"resource,omitempty"`

	// +omitempty
	Authorize *bool `json:"authorize,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
}

type PipelinePermissionStatus struct {
	rtv1.ManagedStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
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
