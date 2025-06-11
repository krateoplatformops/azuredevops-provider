package policies

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"

	policiesv1alpha1 "github.com/krateoplatformops/azuredevops-provider/apis/policies/v1alpha1"
	"github.com/krateoplatformops/azuredevops-provider/internal/clients/azuredevops/policies"
	"github.com/krateoplatformops/azuredevops-provider/internal/resolvers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func isUpdated(ctx context.Context, client client.Client, cr *policiesv1alpha1.Policy, response *policies.PolicyBody) bool {
	if cr.Spec.PolicyBody.IsBlocking != response.IsBlocking {
		return false
	}
	if cr.Spec.PolicyBody.IsDeleted != response.IsDeleted {
		return false
	}
	if cr.Spec.PolicyBody.IsEnabled != response.IsEnabled {
		return false
	}
	if cr.Spec.PolicyBody.IsEnterpriseManaged != response.IsEnterpriseManaged {
		return false
	}
	if cr.Spec.PolicyBody.Type.Id != response.Type.Id {
		return false
	}
	if !compareSettings(ctx, client, cr.Spec.PolicyBody.Settings, response.Settings) {
		return false
	}
	if cr.Spec.PolicyBody.IsBlocking != response.IsBlocking {
		return false
	}

	return true
}

func compareSettings(ctx context.Context, client client.Client, pSpec policiesv1alpha1.PolicySettings, pResponse policies.PolicySettings) bool {
	// Due to the way this API works, we can safely compare only the scope
	if !compareUnorderedScopeArrays(specScopeToClientScope(ctx, client, pSpec.Scope), pResponse.Scope) {
		return false
	}

	// if pSpec.MinimumApproverCount != pResponse.MinimumApproverCount {
	// 	return false
	// }
	// if pSpec.AddedFilesOnly != pResponse.AddedFilesOnly {
	// 	return false
	// }
	// if pSpec.CreatorVoteCounts != pResponse.CreatorVoteCounts {
	// 	return false
	// }
	// if pSpec.EnforceConsistentCase != pResponse.EnforceConsistentCase {
	// 	return false
	// }
	// if pSpec.UseSquashMerge != pResponse.UseSquashMerge {
	// 	return false
	// }
	// if pSpec.UseUncompressedSize != pResponse.UseUncompressedSize {
	// 	return false
	// }
	// if pSpec.ValidDuration != pResponse.ValidDuration {
	// 	return false
	// }
	// if pSpec.BuildDefinitionId != pResponse.BuildDefinitionId {
	// 	return false
	// }
	// if pSpec.MaximumGitBlobSizeInBytes != pResponse.MaximumGitBlobSizeInBytes {
	// 	return false
	// }
	// if pSpec.Message != pResponse.Message {
	// 	return false
	// }
	// if pSpec.EnforceConsistentCase != pResponse.EnforceConsistentCase {
	// 	return false
	// }
	// if compareUnorderedArrays(pSpec.RequiredReviewerIds, pResponse.RequiredReviewerIds) {
	// 	return false
	// }
	// if !compareUnorderedStringArrays(pSpec.FileNamePatterns, pResponse.FileNamePatterns) {
	// 	return false
	// }
	return true
}

func customResourceToPolicy(ctx context.Context, client client.Client, cr *policiesv1alpha1.Policy) (*policies.PolicyBody, error) {
	b, err := json.Marshal(cr.Spec.PolicyBody)
	if err != nil {
		return nil, err
	}

	var pr policies.PolicyBody
	if err := json.Unmarshal(b, &pr); err != nil {
		return nil, err
	}

	pr.Settings.Scope = specScopeToClientScope(ctx, client, cr.Spec.PolicyBody.Settings.Scope)

	return &pr, nil

}

func specScopeToClientScope(ctx context.Context, client client.Client, specScope []policiesv1alpha1.Scope) []policies.Scope {
	var clientScope []policies.Scope

	for _, scope := range specScope {
		repo, err := resolvers.ResolveGitRepository(ctx, client, scope.RepositoryRef)
		if err != nil {
			clientScope = append(clientScope, policies.Scope{
				RefName:   scope.RefName,
				MatchKind: scope.MatchKind,
			})
			continue
		}
		clientScope = append(clientScope, policies.Scope{
			RefName:      scope.RefName,
			MatchKind:    scope.MatchKind,
			RepositoryId: repo.Status.Id,
		})
	}
	return clientScope
}

func compareUnorderedScopeArrays(a, b []policies.Scope) bool {
	if len(a) != len(b) {
		return false
	}

	scopeCount := make(map[policies.Scope]int)
	for _, scope := range a {
		scopeCount[scope]++
	}

	for _, scope := range b {
		if scopeCount[scope] == 0 {
			return false
		}
		scopeCount[scope]--
	}

	for _, count := range scopeCount {
		if count != 0 {
			return false
		}
	}

	return true
}

func compareUnorderedStringArrays(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	return reflect.DeepEqual(a, b)
}

func compareUnorderedArrays(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Ints(a)
	sort.Ints(b)
	return reflect.DeepEqual(a, b)
}
