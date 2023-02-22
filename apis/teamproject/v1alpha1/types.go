package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectorConfigSelectors selects a ConnectorConfig variable.
type ConnectorConfigSelector struct {
	// Name is the name of a connector config.
	Name string `json:"name"`
	// Namespace is the namespace where the connector config belongs.
	Namespace string `json:"namespace"`
}

type Versioncontrol struct {
	// SourceControlType:
	SourceControlType string `json:"sourceControlType"`
}

// ProcessTemplate define reusable content in Azure Devops.
type ProcessTemplate struct {
	// TemplateTypeId: id of the desired template
	TemplateTypeId string `json:"templateTypeId"`
}

// Capabilities this project has
type Capabilities struct {
	Versioncontrol *Versioncontrol `json:"versioncontrol"`

	ProcessTemplate *ProcessTemplate `json:"processTemplate"`
}

// TeamProjectSpec defines the desired state of TeamProject
type TeamProjectSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *ConnectorConfigSelector `json:"connectorConfigRef,omitempty"`

	// Organization: the organization name.
	// +immutable
	Organization string `json:"organization"`

	// OrganizationRef - A reference to an Organization.
	// OrganizationRef *rtv1.Reference `json:"organizationRef,omitempty"`

	// Name: the name of the project.
	// +immutable
	Name string `json:"name"`

	// Description: the project's description (if any).
	// +optional
	Description string `json:"description,omitempty"`

	// Visibility: project visibility: private, public (default: private).
	// +optional
	Visibility *string `json:"visibility,omitempty"`

	// Capabilities: set of capabilities this project has
	// (such as process template & version control).
	Capabilities Capabilities `json:"capabilities,omitempty"`
}

// TeamProjectStatus defines the observed state of a TeamProject
type TeamProjectStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	// Id: project identifier.
	// +optional
	Id string `json:"id,omitempty"`

	// Project revision.
	Revision uint64 `json:"revision,omitempty"`

	// State: the current state of the project..
	// +optional
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// TeamProject is the Schema for the teamprojects API
type TeamProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TeamProjectSpec   `json:"spec,omitempty"`
	Status TeamProjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TeamProjectList contains a list of TeamProject
type TeamProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TeamProject `json:"items"`
}
