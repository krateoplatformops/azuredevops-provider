package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Variable struct {
	IsSecret *bool   `json:"isSecret,omitempty"`
	Value    *string `json:"value,omitempty"`
}

type BuildResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type ContainerResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type PackageResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type PipelineResourceParameters struct {
	Version *string `json:"version,omitempty"`
}

type RepositoryResourceParameters struct {
	RefName *string `json:"refName,omitempty"`
	// This is the security token to use when connecting to the repository.
	Token *string `json:"token,omitempty"`
	// Optional. This is the type of the token given. If not provided, a type of "Bearer" is assumed. Note: Use "Basic" for a PAT token.
	TokenType *string `json:"tokenType,omitempty"`
	Version   *string `json:"version,omitempty"`
}

type RunResourcesParameters struct {
	Builds       map[string]BuildResourceParameters      `json:"builds,omitempty"`
	Containers   map[string]ContainerResourceParameters  `json:"containers,omitempty"`
	Packages     map[string]PackageResourceParameters    `json:"packages,omitempty"`
	Pipelines    map[string]PipelineResourceParameters   `json:"pipelines,omitempty"`
	Repositories map[string]RepositoryResourceParameters `json:"repositories,omitempty"`
}

type RunPipelineParameters struct {
	// If true, don't actually create a new run.
	// Instead, return the final YAML document after parsing templates.
	// +optional
	PreviewRun *bool `json:"previewRun,omitempty"`

	// The resources the run requires.
	// +optional
	Resources *RunResourcesParameters `json:"resources,omitempty"`

	// +optional
	StagesToSkip []string `json:"stagesToSkip,omitempty"`

	// +optional
	TemplateParameters map[string]string `json:"templateParameters,omitempty"`

	// +optional
	Variables map[string]Variable `json:"variables,omitempty"`
	// YamlOverride: If you use the preview run option, you may optionally supply different YAML.
	// This allows you to preview the final YAML document without committing a changed file.
	// +optional
	YamlOverride *string `json:"yamlOverride,omitempty"`
}

// Run defines the desired state of Run
type RunSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// Optional additional parameters for this run.
	// +optional
	RunParameters *RunPipelineParameters `json:"runParameters,omitempty"`

	// PipelineRef: reference to the pipeline.
	PipelineRef *rtv1.Reference `json:"pipelineRef,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
}

type RunStatus struct {
	rtv1.ManagedStatus `json:",inline"`

	// Run ID
	Id *int `json:"id,omitempty"`

	PipelineId *int `json:"pipelineId,omitempty"`

	State *string `json:"state,omitempty"`

	// URL of the Run
	Url *string `json:"url,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// Run is the Schema for the Runs API
type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec   `json:"spec,omitempty"`
	Status RunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RunList contains a list of Run
type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Run `json:"items"`
}
