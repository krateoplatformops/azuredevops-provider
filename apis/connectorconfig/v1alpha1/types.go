package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConnectorConfigSpec struct {
	// ApiUrl: the baseUrl for the REST API provider.
	// +immutable
	ApiUrl string `json:"apiUrl,omitempty"`

	// Credentials required to authenticate ReST API server.
	Credentials *rtv1.CredentialSelectors `json:"credentials"`

	// Verbose is true dumps your client requests and responses.
	// +optional
	//Verbose *bool `json:"verbose,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}

// ConnectorConfigSpec is the Schema for the AzureDevops Client
type ConnectorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConnectorConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// ConnectorConfigList contains a list of ConnectorConfig
type ConnectorConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectorConfig `json:"items"`
}
