package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VariableGroup struct {
	// PrincipalName: the name of the VariableGroup. This field correspond to VariableGroup's email address if the VariableGroup is an Azure Active Directory VariableGroup.
	// +optional
	Name *string `json:"name,omitempty"`

	// OriginID: the origin ID of the VariableGroup. If set, the VariableGroup is assumed to be an Azure Active Directory VariableGroup.
	// +optional
	OriginID *string `json:"originId,omitempty"`
}

// VariableGroups defines the desired state of VariableGroups
type VariableGroupsSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// Name: the name of the VariableGroup. Must be the same as the one in 'variableGroupProjectReferences' due to the limitation of Azure DevOps REST API.
	Name *string `json:"name,omitempty"`

	// Description: the description of the VariableGroup.
	// +optional
	Description *string `json:"description,omitempty"`

	// VariableGroupProjectReferences: a variable group reference is a shallow reference to variable group. Currently, variableGroups cannot be shared across projects so must contain only one project. https://developercommunity.visualstudio.com/t/variablegroup-cannot-be-shared-via-rest-api/488577
	VariableGroupProjectReferences []VariableGroupProjectReference `json:"variableGroupProjectReferences"`

	// Type: the type of the VariableGroup.
	// +optional
	Type *string `json:"type,omitempty"`

	// Variables: a map of variables in the VariableGroup.
	Variables map[string]VariableValue `json:"variables"`
}

type VariableGroupProjectReference struct {
	// Description: Gets or sets description of the variable group.
	// +optional
	Description *string `json:"description,omitempty"`

	// Name: Gets or sets name of the variable group.
	// +optional
	Name *string `json:"name,omitempty"`

	// ProjectReference: Gets or sets project reference of the variable group.
	// +optional
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`
}

type VariableValue struct {
	// IsReadOnly: the flag to indicate whether the variable is read-only.
	// +optional
	IsReadOnly bool `json:"isReadOnly"`

	// Value: the value of the variable.
	// +optional
	Value string `json:"value"`

	// IsSecret: the flag to indicate whether the variable value is secret.
	// +optional
	IsSecret bool `json:"isSecret"`
}

type VariableGroupsStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	Id                 string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// VariableGroups is the Schema for the VariableGroups API
type VariableGroups struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VariableGroupsSpec   `json:"spec,omitempty"`
	Status VariableGroupsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VariableGroupsList contains a list of VariableGroups
type VariableGroupsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VariableGroups `json:"items"`
}
