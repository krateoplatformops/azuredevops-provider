package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConnectorSpec struct {
	// ApiUrl: the baseUrl for the REST API provider.
	// +immutable
	ApiUrl string `json:"apiUrl,omitempty"`

	// Credentials required to authenticate ReST API server.
	Credentials *rtv1.CredentialSelectors `json:"credentials"`

	// Verbose is true dumps your client requests and responses.
	// +optional
	Verbose *bool `json:"verbose,omitempty"`
}

type GitRepositorySpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfig: configuration spec for the REST API client.
	// +immutable
	ConnectorConfig ConnectorSpec `json:"connectorConfig"`

	// Org: organization name.
	// +optional
	Org *string `json:"org,omitempty"`

	// Project: TeamProject name or ID.
	Project string `json:"project,omitempty"`

	// PojectIdRef - A reference to a TeamProject to retrieve its id.
	PojectIdRef *rtv1.Reference `json:"projectIdRef,omitempty"`

	// PojectIdRefSelector - Select a reference to a TeamProject to retrieve its id.
	PojectIdRefSelector *rtv1.Selector `json:"projectIdSelector,omitempty"`

	// Name: name of the Git repository.
	Name string `json:"name,omitempty"`
}

// GitRepositoryStatus defines the observed state of Repository
type GitRepositoryStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	Id            string `json:"id,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	RemoteUrl     string `json:"remoteUrl,omitempty"`
	SshUrl        string `json:"sshUrl,omitempty"`
	Url           string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"

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
