package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpstreamStatus string

const (
	Disabled UpstreamStatus = "disabled"
	Ok       UpstreamStatus = "ok"
)

type UpstreamStatusDetail struct {
	Reason string `json:"reason,omitempty"`
}

// type UpstreamSource struct {
// 	DeletedDate                  string                 `json:"deletedDate,omitempty"`
// 	DisplayLocation              string                 `json:"displayLocation,omitempty"`
// 	ID                           string                 `json:"id,omitempty"`
// 	InternalUpstreamCollectionId string                 `json:"internalUpstreamCollectionId,omitempty"`
// 	InternalUpstreamFeedId       string                 `json:"internalUpstreamFeedId,omitempty"`
// 	InternalUpstreamProjectId    string                 `json:"internalUpstreamProjectId,omitempty"`
// 	InternalUpstreamViewId       string                 `json:"internalUpstreamViewId,omitempty"`
// 	Location                     string                 `json:"location,omitempty"`
// 	Name                         string                 `json:"name,omitempty"`
// 	Protocol                     string                 `json:"protocol,omitempty"`
// 	ServiceEndpointId            string                 `json:"serviceEndpointId,omitempty"`
// 	ServiceEndpointProjectId     string                 `json:"serviceEndpointProjectId,omitempty"`
// 	Status                       UpstreamStatus         `json:"status,omitempty"`
// 	StatusDetails                []UpstreamStatusDetail `json:"statusDetails,omitempty"`
// 	UpstreamSourceType           UpstreamSourceType     `json:"upstreamSourceType,omitempty"`
// }

type UpstreamSource struct {
	// Location: The location of the upstream source.
	Location *string `json:"location,omitempty"`
	// Name: The name of the upstream source.
	Name *string `json:"name,omitempty"`
	// Protocol: The protocol of the upstream source. Possible values are: [NuGet, Npm, Maven, PyPi, Powershell, Docker].
	Protocol *string `json:"protocol,omitempty"`
	// UpstraemSourceType: The type of the upstream source. Possible values are: [internal, public].
	UpstreamSourceType *string `json:"upstreamSourceType,omitempty"`
}

// Feed defines the desired state of Feed
type FeedSpec struct {
	rtv1.ManagedSpec `json:",inline"`

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

	// +omitempty
	IsReadOnly *bool `json:"isReadOnly,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// A list of sources that this feed will fetch packages from. An empty list indicates that this feed will not search any additional sources for packages.
	// UpstreamSources with the same "location" field MUST have the same "name" field.
	// +optional
	UpstreamSources []UpstreamSource `json:"upstreamSources,omitempty"`
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
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url",priority=10
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
