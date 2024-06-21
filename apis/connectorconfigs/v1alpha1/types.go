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

type APIVersionConfig struct {
	// +optional
	CheckConfigurations *string `json:"checkconfiguration,omitempty"`
	// +optional
	Endpoints *string `json:"endpoints,omitempty"`
	// +optional
	Environments *string `json:"environments,omitempty"`
	// +optional
	Feeds *string `json:"feeds,omitempty"`
	// +optional
	FeedPermissions *string `json:"feedpermissions,omitempty"`
	// +optional
	Groups *string `json:"groups,omitempty"`
	// +optional
	Pipelines *string `json:"pipelines,omitempty"`
	// +optional
	PipelinePermissions *string `json:"pipelinepermissions,omitempty"`
	// +optional
	Projects *string `json:"projects,omitempty"`
	// +optional
	PullRequests *string `json:"pullrequests,omitempty"`
	// +optional
	Queues *string `json:"queues,omitempty"`
	// +optional
	Repositories *string `json:"repositories,omitempty"`
	// +optional
	RepositoryPermissions *string `json:"repositorypermissions,omitempty"`
	// +optional
	Runs *string `json:"runs,omitempty"`
	// +optional
	SecureFiles *string `json:"securefiles,omitempty"`
	// +optional
	Teams *string `json:"teams,omitempty"`
	// +optional
	Users *string `json:"users,omitempty"`
	// +optional
	VariableGroups *string `json:"variablegroups,omitempty"`
	// +optional
	Descriptors *string `json:"descriptors,omitempty"`
	// +optional
	Memberships *string `json:"memberships,omitempty"`
	// +optional
	Identities *string `json:"identities,omitempty"`
	// +optional
	Pools *string `json:"pools,omitempty"`
	// +optional
	Definitions *string `json:"definitions,omitempty"`
	// +optional
	Operations *string `json:"operations,omitempty"`
	// +optional
	Policies *string `json:"policies,omitempty"`
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

	// APIVersionConfig: the API version configuration.
	// +optional
	APIVersionConfig *APIVersionConfig `json:"apiVersionConfig,omitempty"`
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
