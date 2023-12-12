package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Membership struct {
	// Used if ProjectRef is not set
	// +optional
	Organization *string `json:"organization"`
	// Reference to the project
	// +optional
	ProjectRef *rtv1.Reference `json:"projectRef"`
}
type GroupIdentifier struct {
	// GroupsName: name of the group
	// +optional
	GroupsName string `json:"groupName"`
	// OriginID: the origin ID of the user.
	// +optional
	OriginID string `json:"originId,omitempty"`
}

// Groups defines the desired state of Groups
type GroupsSpec struct {
	rtv1.ManagedSpec `json:",inline"`
	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// Membership: membership of the group - One of Organization or ProjectRef must be set
	// +required
	Membership Membership `json:"membership"`

	// // GroupsName: name of the group
	// // +required
	// GroupsName string `json:"groupName"`

	// One of origidId or groupName must be specified
	// +required
	GroupIdentifier `json:",inline"`

	// Description: description of the group
	// +optional
	Description string `json:"description,omitempty"`

	// GroupRefs: the groups to which the group belongs.
	// +optional
	GroupsRefs []rtv1.Reference `json:"groupRefs"`
}

type GroupsStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	Descriptor         *string `json:"descriptor,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Groups is the Schema for the Groups API
type Groups struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupsSpec   `json:"spec,omitempty"`
	Status GroupsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GroupsList contains a list of Groups
type GroupsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Groups `json:"items"`
}
