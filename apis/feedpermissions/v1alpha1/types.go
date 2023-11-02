package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UserResource struct {
	// The user identity type to add
	// [build-service]
	Type *string `json:"type"`
	// The role for this identity on a feed.
	// [custom, none, reader, contributor, administrator, collaborator]
	Role *string `json:"role"`
	// ProjectRef - A reference to the teamproject that owns the user
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`
}

// FeedPermission defines the desired state of FeedPermission
type FeedPermissionSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ProjectRef - A reference to a TeamProject that owns the feed.
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// Name or ID of the feed
	Feed *string `json:"feed,omitempty"`

	// User Permissions and type
	User *UserResource `json:"user,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
}

type FeedPermissionStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	IdentityDescriptor string `json:"identityDescriptor,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="DESCRIPTOR",type="string",JSONPath=".status.identityDescriptor"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// FeedPermission is the Schema for the FeedPermissions API
type FeedPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeedPermissionSpec   `json:"spec,omitempty"`
	Status FeedPermissionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FeedPermissionList contains a list of FeedPermission
type FeedPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FeedPermission `json:"items"`
}
