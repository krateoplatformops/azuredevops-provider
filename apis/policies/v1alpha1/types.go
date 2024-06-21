package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	resouce "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyType struct {
	// Display name of the policy type.
	// +optional
	DisplayName string `json:"displayName"`
	// The policy type ID.
	Id string `json:"id"`
}
type Scope struct {
	RefName   string `json:"refName"`
	MatchKind string `json:"matchKind"`
	// RepositoryRef: reference to an existing CR of a repository.
	// +optional
	RepositoryRef *rtv1.Reference `json:"repositoryRef,omitempty"`
}

type PolicySettings struct {
	// +optional
	MinimumApproverCount int `json:"minimumApproverCount"`
	// +optional
	CreatorVoteCounts bool `json:"creatorVoteCounts"`
	// +optional
	Scope []Scope `json:"scope"`
	// +optional
	BuildDefinitionId int `json:"buildDefinitionId"`
	// +optional
	RequiredReviewerIds []int `json:"requiredReviewerIds"`
	// +optional
	FileNamePatterns []string `json:"fileNamePatterns"`
	// +optional
	AddedFilesOnly bool `json:"addedFilesOnly"`
	// +optional
	Message string `json:"message"`
	// +optional
	EnforceConsistentCase bool `json:"enforceConsistentCase"`
	// +optional
	MaximumGitBlobSizeInBytes int `json:"maximumGitBlobSizeInBytes"`
	// +optional
	UseUncompressedSize bool `json:"useUncompressedSize"`
	// +optional
	UseSquashMerge bool `json:"useSquashMerge"`
	// +optional
	ManualQueueOnly bool `json:"manualQueueOnly"`
	// +optional
	QueueOnSourceUpdateOnly bool `json:"queueOnSourceUpdateOnly"`
	// +optional
	ValidDuration resouce.Quantity `json:"validDuration"`
}

type PolicyBody struct {
	// ProjectRef - A reference to a TeamProject.
	// +optional
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// Type - The policy configuration type.
	// +optional
	Type PolicyType `json:"type,omitempty"`

	// IsBlocking - Indicates whether the policy is blocking.
	IsBlocking bool `json:"isBlocking"`

	// IsEnabled - Indicates whether the policy is enabled.
	// +optional
	IsEnabled bool `json:"isEnabled"`

	// IsEnterpriseManaged - If set, this policy requires "Manage Enterprise Policies" permission to create, edit, or delete.
	// +optional
	IsEnterpriseManaged bool `json:"isEnterpriseManaged"`

	// IsDeleted - Indicates whether the policy has been (soft) deleted.
	// +optional
	IsDeleted bool `json:"isDeleted"`

	// Settings - The policy configuration settings. Only 'settings.scope' is compared when checking for configuration drift due to the api undocumented behavior.
	// +optional
	Settings PolicySettings `json:"settings,omitempty"`

	// ID - The policy configuration ID. You can specify this field when you need to retrieve or update an existing policy configuration.
	// +optional
	ID *int `json:"id,omitempty"`
}

// Policy defines the desired state of Policy
type PolicySpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// PolicyBody - The policy configuration.
	// +required
	PolicyBody PolicyBody `json:"policyBody"`
}

type PolicyStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	// ID - The policy configuration ID.
	// +optional
	ID *int `json:"id,omitempty"`

	// URL - The URL where the policy configuration can be retrieved.
	// +optional
	URL *string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url",priority=10
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Policy is the Schema for the Policys API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
