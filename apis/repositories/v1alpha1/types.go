package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitRepositorySpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// Organization: the organization name.
	//Organization string `json:"organization"`

	// Project: TeamProject name or ID.
	// +optional
	Project *string `json:"project,omitempty"`

	// PojectRef - A reference to a TeamProject.
	PojectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// Name: name of the Git repository.
	Name string `json:"name,omitempty"`

	// Init: initialize the Git repository.
	Initialize *bool `json:"initialize,omitempty"`
}

// GitRepositoryStatus defines the observed state of Repository
type GitRepositoryStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	Id            string `json:"id,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	SshUrl        string `json:"sshUrl,omitempty"`
	Url           string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id",priority=10
//+kubebuilder:printcolumn:name="SSH_URL",type="string",JSONPath=".status.sshUrl"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// GitRepository is the Schema for the gitrepository API
type GitRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitRepositorySpec   `json:"spec,omitempty"`
	Status GitRepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitRepositoryList contains a list of GitRepository
type GitRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitRepository `json:"items"`
}
