package pullrequests

import (
	"encoding/json"
	"reflect"
	"strings"

	pullrequestsv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/pullrequests/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/pullrequests"
)

func createPRWithModifiedFields(cr *pullrequestsv1alpha1.PullRequest, response *pullrequests.PullRequest) (pr *pullrequests.PullRequest, err error) {
	pr = &pullrequests.PullRequest{}
	if cr.Spec.PullRequest.Title != response.Title {
		pr.Title = cr.Spec.PullRequest.Title
	}
	if cr.Spec.PullRequest.Description != response.Description {
		pr.Description = cr.Spec.PullRequest.Description
	}
	if cr.Spec.PullRequest.Status != "" && !strings.EqualFold(cr.Spec.PullRequest.Status, response.Status) {
		pr.Status = cr.Spec.PullRequest.Status
	}
	if cr.Spec.PullRequest.TargetRefName != response.TargetRefName {
		pr.TargetRefName = cr.Spec.PullRequest.TargetRefName
	}
	if response.CompletionOptions != nil && !compareCompletionOptions(cr, response) {
		pr.CompletionOptions = &pullrequests.CompletionOptions{
			DeleteSourceBranch:          cr.Spec.PullRequest.CompletionOptions.DeleteSourceBranch,
			MergeCommitMessage:          cr.Spec.PullRequest.CompletionOptions.MergeCommitMessage,
			SquashMerge:                 cr.Spec.PullRequest.CompletionOptions.SquashMerge,
			TransitionWorkItems:         cr.Spec.PullRequest.CompletionOptions.TransitionWorkItems,
			TriggeredByAutoComplete:     cr.Spec.PullRequest.CompletionOptions.TriggeredByAutoComplete,
			AutoCompleteIgnoreConfigIds: cr.Spec.PullRequest.CompletionOptions.AutoCompleteIgnoreConfigIds,
			BypassPolicy:                cr.Spec.PullRequest.CompletionOptions.BypassPolicy,
			BypassReason:                cr.Spec.PullRequest.CompletionOptions.BypassReason,
			MergeStrategy:               cr.Spec.PullRequest.CompletionOptions.MergeStrategy,
		}
	}
	if response.MergeOptions != nil && !compareMergeOptions(cr, response) {
		pr.MergeOptions = &pullrequests.GitPullRequestMergeOptions{
			SquashMerge:        cr.Spec.PullRequest.MergeOptions.SquashMerge,
			CreateMergeCommit:  cr.Spec.PullRequest.MergeOptions.CreateMergeCommit,
			MergeCommitMessage: cr.Spec.PullRequest.MergeOptions.MergeCommitMessage,
			MergeStrategy:      cr.Spec.PullRequest.MergeOptions.MergeStrategy,
		}

	}
	if response.AutoCompleteSetBy != nil && cr.Spec.PullRequest.AutoCompleteSetBy.Id != response.AutoCompleteSetBy.Id {
		pr.AutoCompleteSetBy = &pullrequests.IdentityRef{
			Id: cr.Spec.PullRequest.AutoCompleteSetBy.Id,
		}
	}
	return pr, nil
}

func compareCompletionOptions(cr *pullrequestsv1alpha1.PullRequest, response *pullrequests.PullRequest) bool {
	if cr.Spec.PullRequest.CompletionOptions.BypassPolicy != response.CompletionOptions.BypassPolicy {
		return false
	}
	if cr.Spec.PullRequest.CompletionOptions.BypassReason != response.CompletionOptions.BypassReason {
		return false
	}
	if cr.Spec.PullRequest.CompletionOptions.DeleteSourceBranch != response.CompletionOptions.DeleteSourceBranch {
		return false
	}
	if cr.Spec.PullRequest.CompletionOptions.MergeCommitMessage != response.CompletionOptions.MergeCommitMessage {
		return false
	}
	if response.CompletionOptions.MergeStrategy != "noFastForward" && cr.Spec.PullRequest.CompletionOptions.MergeStrategy != response.CompletionOptions.MergeStrategy {
		return false
	}
	if cr.Spec.PullRequest.CompletionOptions.SquashMerge != response.CompletionOptions.SquashMerge {
		return false
	}
	if cr.Spec.PullRequest.CompletionOptions.TransitionWorkItems != response.CompletionOptions.TransitionWorkItems {
		return false
	}
	if cr.Spec.PullRequest.CompletionOptions.TriggeredByAutoComplete != response.CompletionOptions.TriggeredByAutoComplete {
		return false
	}
	if !reflect.DeepEqual(cr.Spec.PullRequest.CompletionOptions.AutoCompleteIgnoreConfigIds, response.CompletionOptions.AutoCompleteIgnoreConfigIds) {
		return false
	}
	return true
}
func compareMergeOptions(cr *pullrequestsv1alpha1.PullRequest, response *pullrequests.PullRequest) bool {
	if cr.Spec.PullRequest.MergeOptions.SquashMerge != response.MergeOptions.SquashMerge {
		return false
	}
	if cr.Spec.PullRequest.MergeOptions.CreateMergeCommit != response.MergeOptions.CreateMergeCommit {
		return false
	}
	if cr.Spec.PullRequest.MergeOptions.MergeCommitMessage != response.MergeOptions.MergeCommitMessage {
		return false
	}
	if cr.Spec.PullRequest.MergeOptions.MergeStrategy != response.MergeOptions.MergeStrategy {
		return false
	}
	return true

}

func isUpdated(cr *pullrequestsv1alpha1.PullRequest, response *pullrequests.PullRequest) bool {
	if cr.Spec.PullRequest.Title != response.Title {
		return false
	}
	if cr.Spec.PullRequest.Description != response.Description {
		return false
	}
	if cr.Spec.PullRequest.Status != "" && !strings.EqualFold(cr.Spec.PullRequest.Status, response.Status) {
		return false
	}
	if cr.Spec.PullRequest.TargetRefName != response.TargetRefName {
		return false
	}
	if response.CompletionOptions != nil && !compareCompletionOptions(cr, response) {
		return false
	}
	if response.MergeOptions != nil && !compareMergeOptions(cr, response) {
		return false
	}
	if response.AutoCompleteSetBy != nil && cr.Spec.PullRequest.AutoCompleteSetBy.Id != response.AutoCompleteSetBy.Id {
		return false
	}
	return true
}

func customResourceToPullRequest(cr *pullrequestsv1alpha1.PullRequest) (*pullrequests.PullRequest, error) {
	b, err := json.Marshal(cr.Spec.PullRequest)
	if err != nil {
		return nil, err
	}

	var pr pullrequests.PullRequest
	if err := json.Unmarshal(b, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}
