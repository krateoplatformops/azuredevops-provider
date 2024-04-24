package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecureFiles defines the desired state of SecureFiles
type SecureFilesSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// ProjectRef: the project to which the SecureFile belongs.
	// +required
	ProjectRef *rtv1.Reference `json:"projectRef"`

	// Name: the name of the SecureFile.
	// +required
	Name string `json:"name"`
}

type SecureFilesStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	Id                 *string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// SecureFiles is the Schema for the SecureFiles API
type SecureFiles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecureFilesSpec   `json:"spec,omitempty"`
	Status SecureFilesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecureFilesList contains a list of SecureFiles
type SecureFilesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecureFiles `json:"items"`
}
