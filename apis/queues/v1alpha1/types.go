package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QueueSpec defines the desired state of Queue
type QueueSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// Organization
	// +optional
	Organization *string `json:"organization,omitempty"`

	// Project: TeamProject name or ID.
	// +optional
	Project *string `json:"project,omitempty"`

	// ProjectRef - A reference to a TeamProject.
	// +optional
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// Name of the feed.
	// +optional
	Name *string `json:"name,omitempty"`

	// Pool Name
	// +immutable
	Pool string `json:"pool"`
}

// QueueStatus defines the observed state of a Queue
type QueueStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	// Id: project identifier.
	// +optional
	Id *int `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Queue is the Schema for the teamprojects API
type Queue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueSpec   `json:"spec,omitempty"`
	Status QueueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QueueList contains a list of Queue
type QueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Queue `json:"items"`
}
