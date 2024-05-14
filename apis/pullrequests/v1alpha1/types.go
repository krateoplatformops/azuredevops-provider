package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitPullRequest struct {
	// Specifies the options for completing the pull request.
	// +optional
	CompletionOptions CompletionOptions `json:"completionOptions,omitempty"`
	// If set, auto-complete is enabled for this pull request and this is the identity that enabled it.
	// +optional
	AutoCompleteSetBy IdentityRef `json:"autoCompleteSetBy,omitempty"`
	// ID of the code review associated with this pull request.
	// +optional
	CodeReviewId int `json:"codeReviewId,omitempty"`
	// The user who closed the pull request.
	// +optional
	ClosedBy IdentityRef `json:"closedBy,omitempty"`
	// The ID of the commit this pull request is based on.
	// +optional
	CommitId string `json:"commitId,omitempty"`
	// Description of the pull request.
	// +optional
	Description string `json:"description,omitempty"`
	// Indicates if the pull request is a draft.
	// +optional
	IsDraft bool `json:"isDraft,omitempty"`
	// ID of the pull request merge.
	// +optional
	MergeId string `json:"mergeId,omitempty"`
	// The merge status of the pull request.
	// +optional
	MergeStatus string `json:"mergeStatus,omitempty"`
	// Specifies the options for merging the pull request.
	// +optional
	MergeOptions GitPullRequestMergeOptions `json:"mergeOptions,omitempty"`
	// The project associated with this pull request.
	// +optional
	Project TeamProjectReference `json:"project,omitempty"`
	// The resource version.
	// +optional
	ResourceVersion int `json:"resourceVersion,omitempty"`
	// List of reviewers on the pull request and the vote on the pull request.
	// +optional
	Reviewers []IdentityRefWithVote `json:"reviewers,omitempty"`
	// The status of the pull request.
	// +optional
	Status string `json:"status,omitempty"`
	// The source reference name of the pull request.
	// +optional
	SourceRefName string `json:"sourceRefName,omitempty"`
	// The target reference name of the pull request.
	// +optional
	TargetRefName string `json:"targetRefName,omitempty"`
	// Title of the pull request.
	// +optional
	Title string `json:"title,omitempty"`
	// URL of the pull request.
	// +optional
	Url string `json:"url,omitempty"`
	// Indicates whether the project supports iterations.
	// +optional
	SupportsIterations bool `json:"supportsIterations,omitempty"`
	// The last commit information associated with the pull request merge.
	// +optional
	LastMergeCommit GitCommitRef `json:"lastMergeCommit,omitempty"`
	// The last source commit associated with the pull request merge.
	// +optional
	LastMergeSourceCommit GitCommitRef `json:"lastMergeSourceCommit,omitempty"`
	// The last target commit associated with the pull request merge.
	// +optional
	LastMergeTargetCommit GitCommitRef `json:"lastMergeTargetCommit,omitempty"`
	// The last merge associated with the pull request.
	// +optional
	LastMerge GitPullRequestMerge `json:"lastMerge,omitempty"`
	// The user who created the pull request.
	// +optional
	CreatedBy IdentityRef `json:"createdBy,omitempty"`
	// List of work item IDs associated with the pull request.
	// +optional
	WorkItemRefs []int `json:"workItemRefs,omitempty"`
	// List of labels associated with the pull request.
	// +optional
	Labels []WebApiTagDefinition `json:"labels,omitempty"`
}

type WebApiTagDefinition struct {
	Active bool   `json:"active,omitempty"`
	Id     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Url    string `json:"url,omitempty"`
}

type IdentityRef struct {
	Id          string `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UniqueName  string `json:"uniqueName,omitempty"`
	Url         string `json:"url,omitempty"`
	ImageUrl    string `json:"imageUrl,omitempty"`
	Descriptor  string `json:"descriptor,omitempty"`
}

type IdentityRefWithVote struct {
	IdentityRef `json:",inline"`
	Vote        int `json:"vote,omitempty"`
}

type GitProject struct {
	Id    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Url   string `json:"url,omitempty"`
	State string `json:"state,omitempty"`
}

type GitCommitRef struct {
	CommitId string `json:"commitId,omitempty"`
	Url      string `json:"url,omitempty"`
}

type GitPullRequestMergeOptions struct {
	SquashMerge        bool   `json:"squashMerge,omitempty"`
	CreateMergeCommit  bool   `json:"createMergeCommit,omitempty"`
	MergeCommitMessage string `json:"mergeCommitMessage,omitempty"`
	MergeStrategy      string `json:"mergeStrategy,omitempty"`
}

type GitPullRequestMerge struct {
	MergeType     string `json:"mergeType,omitempty"`
	MergeCommitId string `json:"mergeCommitId,omitempty"`
	MergeCommit   string `json:"mergeCommit,omitempty"`
}

type TeamProjectReference struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

type CompletionOptions struct {
	BypassPolicy                bool   `json:"bypassPolicy,omitempty"`
	BypassReason                string `json:"bypassReason,omitempty"`
	DeleteSourceBranch          bool   `json:"deleteSourceBranch,omitempty"`
	MergeCommitMessage          string `json:"mergeCommitMessage,omitempty"`
	MergeStrategy               string `json:"mergeStrategy,omitempty"`
	SquashMerge                 bool   `json:"squaredMerge,omitempty"`
	TransitionWorkItems         bool   `json:"transitionWorkItems,omitempty"`
	TriggeredByAutoComplete     bool   `json:"triggeredByAutoComplete,omitempty"`
	AutoCompleteIgnoreConfigIds []int  `json:"autoCompleteIgnoreConfigIds,omitempty"`
}

// PullRequestSpec defines the desired state of PullRequest
type PullRequestSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`

	// ProjectRef: reference to an existing CR of a project.
	// +required
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`

	// RepositoryRef: reference to an existing CR of a repository.
	// +required
	RepositoryRef *rtv1.Reference `json:"repositoryRef,omitempty"`

	PullRequest GitPullRequest `json:"pullRequest,omitempty"`
}

// PullRequestStatus defines the observed state of a PullRequest
type PullRequestStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	Id                 *string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10

// PullRequest is the Schema for the PullRequests API
type PullRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PullRequestSpec   `json:"spec,omitempty"`
	Status PullRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PullRequestList contains a list of PullRequest
type PullRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PullRequest `json:"items"`
}
