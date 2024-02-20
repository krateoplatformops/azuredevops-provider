package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type User struct {
	// PrincipalName: the name of the user. This field correspond to user's email address if the user is an Azure Active Directory user.
	// +optional
	Name *string `json:"name,omitempty"`

	// OriginID: the origin ID of the user. If set, the user is assumed to be an Azure Active Directory user.
	// +optional
	OriginID *string `json:"originId,omitempty"`
}

// Users defines the desired state of Users
type UsersSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// Organization: the organization to which the user belongs.
	// +required
	Organization string `json:"organization"`

	// User: the user to be created or retrieved. Either name or originId must be specified.
	// +required
	User User `json:"user"`

	// GroupRefs: the groups to which the user belongs.
	// +optional
	GroupsRefs []rtv1.Reference `json:"groupRefs"`

	// TeamRefs: the teams to which the user belongs.
	// +optional
	TeamsRefs []rtv1.Reference `json:"teamRefs"`
}

type UsersStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	Descriptor         *string `json:"descriptor,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="DESCRIPTOR",type="string",JSONPath=".status.descriptor"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Users is the Schema for the Users API
type Users struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UsersSpec   `json:"spec,omitempty"`
	Status UsersStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UsersList contains a list of Users
type UsersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Users `json:"items"`
}
