package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	// Type: type of resource.
	// +kubebuilder:validation:Environment,Queue
	// +required
	Type string `json:"type"`
	// Reference: reference to the resource.
	// +required
	ResourceRef *rtv1.Reference `json:"resourceRef"`
}

type Approver struct {
	// ID: approver ID.
	// +optional
	ID *string `json:"id,omitempty"`
	// ApproverRef: approver reference.
	// +optional
	ApproverRef *rtv1.Reference `json:"approverRef,omitempty"`
}

type ApprovalSettings struct {
	// Approvers: list of approvers. id or approverRef must be specified.
	// +optional
	Approvers []Approver `json:"approvers"`
	// ExecutionOrder: execution order of the approvers.
	// +optional
	ExecutionOrder string `json:"executionOrder,omitempty"`
	// Instructions: instructions for approvers.
	// +optional
	Instructions string `json:"instructions,omitempty"`
	// MinRequiredApprovers: minimum number of approvers.
	// +optional
	MinRequiredApprovers int `json:"minRequiredApprovers,omitempty"`
	// BlockedApprovers: list of blocked approvers.
	// +optional
	BlockedApprovers []string `json:"blockedApprovers,omitempty"`
	// RequesterCannotBeApprover: requester cannot be approver.
	// +optional
	RequesterCannotBeApprover bool `json:"requesterCannotBeApprover,omitempty"`
}
type DefinitionRef struct {
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// Refers here for docs https://stackoverflow.com/questions/61471634/add-remove-pipeline-checks-using-rest-api
type TaskCheckSettings struct {
	// Inputs in inline JSON format
	Inputs              string        `json:"inputs,omitempty"`
	LinkedVariableGroup string        `json:"linkedVariableGroup,omitempty"`
	RetryInterval       int           `json:"retryInterval,omitempty"`
	DisplayName         string        `json:"displayName,omitempty"`
	DefinitionRef       DefinitionRef `json:"definitionRef,omitempty"`
}

type ExtendsCheckSettings struct {
	RepositoryType string `json:"repositoryType,omitempty"`
	RepositoryName string `json:"repositoryName,omitempty"`
	RepositoryRef  string `json:"repositoryRef,omitempty"`
	TemplatePath   string `json:"templatePath,omitempty"`
}

// CheckConfiguration defines the desired state of CheckConfiguration
type CheckConfigurationSpec struct {
	rtv1.ManagedSpec `json:",inline"`
	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
	// ProjectRef: project reference.
	// +required
	ProjectRef *rtv1.Reference `json:"projectRef"`
	// Type: type of check configuration.
	// +kubebuilder:validation:Enum=Approval;Task Check;Extends Check
	Type string `json:"type"`
	// Resource is the resource to check.
	// +required
	Resource Resource `json:"resource"`
	// Timeout: timeout in minutes.
	// +required
	Timeout int `json:"timeout"`
	// ApprovalSettings: settings for the check configuration. Only used if type is Approval. If type is Approval, then this field is required.
	// +optional
	ApprovalSettings ApprovalSettings `json:"approvalSettings"`
	// TaskCheckSettings: settings for the check configuration. Only used if type is TaskCheck. If type is TaskCheck, then this field is required.
	// +optional
	TaskCheckSettings TaskCheckSettings `json:"taskCheckSettings"`
	// ExtendsCheckSettings: settings for the check configuration. Only used if type is ExtendsCheck. If type is ExtendsCheck, then this field is required.
	// +optional
	ExtendsCheckSettings []ExtendsCheckSettings `json:"extendsCheckSettings"`
}

type CheckConfigurationStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	// Id: check configuration ID.
	ID *string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// CheckConfiguration is the Schema for the CheckConfigurations API
type CheckConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CheckConfigurationSpec   `json:"spec,omitempty"`
	Status CheckConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CheckConfigurationList contains a list of CheckConfiguration
type CheckConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CheckConfiguration `json:"items"`
}
