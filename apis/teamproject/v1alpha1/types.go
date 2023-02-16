package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TeamProjectSpec defines the desired state of TeamProject
type TeamProjectSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ApiUrl: the baseUrl for the REST API provider.
	// +immutable
	ApiUrl string `json:"apiUrl,omitempty"`

	// Credentials required to authenticate ReST API server.
	Credentials *rtv1.CredentialSelectors `json:"credentials"`

	// Verbose is true dumps your client requests and responses.
	// +optional
	Verbose *bool `json:"verbose,omitempty"`

	// Org: the organization name.
	// +immutable
	Org string `json:"org"`

	// Name: the name of the project.
	// +immutable
	Name string `json:"name"`

	// Description: the project's description (if any).
	// +optional
	Description string `json:"description,omitempty"`

	// Private: the project is only visible to users with explicit access. (default: true).
	// +optional
	Private bool `json:"private,omitempty"`

	// Capabilities: set of capabilities this project has
	// (such as process template & version control).
	// +optional
	Capabilities map[string]map[string]string `json:"capabilities,omitempty"`
}

// TeamProjectStatus defines the observed state of Repo
type TeamProjectStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	// Id: project identifier.
	// +optional
	Id *string `json:"id,omitempty"`

	// Name: project name.
	// +optional
	Name *string `json:"name,omitempty"`

	// State: the current state of the project..
	// +optional
	State *string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"

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
