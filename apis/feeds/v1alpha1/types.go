package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Feed defines the desired state of Feed
type FeedSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// Organization
	// +optional
	Organization *string `json:"organization,omitempty"`

	// Project: TeamProject name or ID.
	// +optional
	Project *string `json:"project,omitempty"`

	// PojectRef - A reference to a TeamProject.
	// +optional
	PojectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// Name of the feed.
	// +optional
	Name *string `json:"name,omitempty"`

	// +omitempty
	IsReadOnly *bool `json:"isReadOnly,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
}

type FeedStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	Id *string `json:"id,omitempty"`

	Url *string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Feed is the Schema for the Feeds API
type Feed struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeedSpec   `json:"spec,omitempty"`
	Status FeedStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeedList contains a list of Feed
type FeedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Feed `json:"items"`
}
