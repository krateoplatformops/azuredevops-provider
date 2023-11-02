package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EndpointAuthorizationParams struct {
	Tenantid                  *string `json:"tenantid,omitempty"`
	ServiceprincipalId        *string `json:"serviceprincipalId,omitempty"`
	AuthenticationType        *string `json:"authenticationType,omitempty"`
	ServiceprincipalKey       *string `json:"serviceprincipalKey,omitempty"`
	Scope                     *string `json:"scope,omitempty"`
	ServiceAccountCertificate *string `json:"serviceAccountCertificate,omitempty"`
	IsCreatedFromSecretYaml   *string `json:"isCreatedFromSecretYaml,omitempty"`
	Apitoken                  *string `json:"apitoken,omitempty"`
}

// Represents the authorization used for service endpoint.
type EndpointAuthorization struct {
	// Gets or sets the parameters for the selected authorization scheme.
	Parameters *EndpointAuthorizationParams `json:"parameters,omitempty"`
	// Gets or sets the scheme used for service endpoint authentication.
	Scheme *string `json:"scheme,omitempty"`
}

type Data struct {
	Environment          *string `json:"environment,omitempty"`
	ScopeLevel           *string `json:"scopeLevel,omitempty"`
	SubscriptionId       *string `json:"subscriptionId,omitempty"`
	SubscriptionName     *string `json:"subscriptionName,omitempty"`
	CreationMode         *string `json:"creationMode,omitempty"`
	AuthorizationType    *string `json:"authorizationType,omitempty"`
	AcceptUntrustedCerts *string `json:"acceptUntrustedCerts,omitempty"`
}

type ProjectReference struct {
	Id   *string `json:"id,omitempty"`
	Name string  `json:"name,omitempty"`
}

type ServiceEndpointProjectReference struct {
	// Gets or sets description of the service endpoint.
	Description *string `json:"description,omitempty"`
	// Gets or sets name of the service endpoint.
	Name *string `json:"name,omitempty"`
	// Gets or sets project reference of the service endpoint.
	ProjectReference *ProjectReference `json:"projectReference,omitempty"`
}

// EndpointSpec defines the desired state of Endpoint
type EndpointSpec struct {
	rtv1.ManagedSpec `json:",inline"`
	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
	// Organization
	// +optional
	Organization *string `json:"organization,omitempty"`
	// Project: TeamProject name or ID.
	// +optional
	Project *string `json:"project,omitempty"`
	// ProjectRef - A reference to a TeamProject.
	// +optional
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`
	// Authorization: the authorization data for talking to the endpoint.
	// +optional
	Authorization *EndpointAuthorization `json:"authorization,omitempty"`
	// Data:
	// +optional
	Data *Data `json:"data,omitempty"`
	// Indicates whether service endpoint is shared with other projects or not.
	// +optional
	IsShared *bool `json:"isShared,omitempty"`
	// Name: the friendly name of the endpoint.
	// +optional
	Name *string `json:"name,omitempty"`
	// Description: the friendly description of the endpoint.
	// +optional
	Description *string `json:"description,omitempty"`
	// Owner of the endpoint Supported values are "library", "agentcloud"
	// +optional
	Owner *string `json:"owner,omitempty"`
	// Type: the type of the endpoint.
	// +optional
	Type *string `json:"type,omitempty"`
	// Url: the url of the endpoint.
	// +optional
	Url *string `json:"url,omitempty"`
	// All other project references where the service endpoint is shared.
	// +optional
	ServiceEndpointProjectReferences []ServiceEndpointProjectReference `json:"serviceEndpointProjectReferences,omitempty"`
}

// EndpointStatus defines the observed state of a Endpoint
type EndpointStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	// Id:
	// +optional
	Id *string `json:"id,omitempty"`

	// Url: the url of the endpoint.
	// +optional
	Url *string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Endpoint is the Schema for the teamprojects API
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointSpec   `json:"spec,omitempty"`
	Status EndpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Endpoint `json:"items"`
}
