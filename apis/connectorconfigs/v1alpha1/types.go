package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A Reference to a named object.
type Reference struct {
	// Name of the referenced object.
	Name string `json:"name"`

	// Namespace of the referenced object.
	Namespace string `json:"namespace"`
}

type ApiUrl struct {
	// Default: the baseUrl for the REST API provider.
	Defautl string `json:"default,omitempty"`
	// Feeds: the baseUrl for the REST API provider.
	Feeds string `json:"feeds,omitempty"`
	// Vssps: the baseUrl for the REST API provider.
	Vssps string `json:"vssps,omitempty"`
}

type ConnectorConfigSpec struct {
	// DEPRECATED: This field is deprecated and will be removed in a future version. Use the ApiUrls field instead.
	// ApiUrl: the baseUrl for the REST API provider.
	// +immutable
	// +optional
	ApiUrl string `json:"apiUrl,omitempty"`

	// ApiUrls: the baseUrl for the REST API provider.
	// +immutable
	// +optional
	ApiUrls *ApiUrl `json:"apiUrls,omitempty"`

	// Credentials required to authenticate ReST API server.
	// +required
	Credentials *rtv1.CredentialSelectors `json:"credentials"`
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
