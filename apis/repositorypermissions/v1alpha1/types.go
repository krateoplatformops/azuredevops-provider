package v1alpha1

import (
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Identity struct {
	// The user identity type to add
	// [build-service, azure-group]
	// +required
	Type *string `json:"type,omitempty"`
	// ProjectRef - The reference to the project that contains the group.
	// +required
	ProjectRef *rtv1.Reference `json:"projectRef,omitempty"`
	// Name - The name of the group. This parameter is ignored if Type is build-service.
	Name *string `json:"name,omitempty"`
}

type Permissions struct {
	// Identity - The identity who has these permissions.
	Identity *Identity `json:"identity,omitempty"`

	// Merge
	// If true, the permissions are added to existing permissions and only the setted perms are mantained by the controller.
	// If false, the permissions are the only permissions for this identity and are mantained exaclty as the CR by the controller.
	// +required
	Merge bool `json:"merge,omitempty"`

	// AllowList - The permissions that this identity has.
	// Possible (case insensitive) values are [administerpermission,genericread,genericcontribute,forcepush,createbranch,createtag,managenote,policyexempt,createrepository,deleterepository,renamerepository,editpolicies,removeotherslocks,managepermissions,pullrequestcontribute,pullrequestbypasspolicy,viewadvsecalerts,dismissadvsecalerts,manageadvsecscanning]
	AllowList []string `json:"allowList,omitempty"`

	// DenyList - The permissions that this identity is explicitly denied.
	// Possible values are [administerpermission,genericread,genericcontribute,forcepush,createbranch,createtag,managenote,policyexempt,createrepository,deleterepository,renamerepository,editpolicies,removeotherslocks,managepermissions,pullrequestcontribute,pullrequestbypasspolicy,viewadvsecalerts,dismissadvsecalerts,manageadvsecscanning]
	DenyList []string `json:"denyList,omitempty"`
}

// RepositoryPermission defines the desired state of RepositoryPermission
type RepositoryPermissionSpec struct {
	rtv1.ManagedSpec `json:",inline"`

	// RepositoryRef - The reference to the repository.
	// +required
	RepositoryRef *rtv1.Reference `json:"repositoryRef,omitempty"`

	// Permissions - The permissions to set.
	// +required
	Permissions *Permissions `json:"permissions,omitempty"`

	// ConnectorConfigRef: configuration spec for the REST API client.
	// +immutable
	ConnectorConfigRef *rtv1.Reference `json:"connectorConfigRef,omitempty"`
}

type RepositoryPermissionStatus struct {
	rtv1.ManagedStatus `json:",inline"`
	IdentityDescriptor string `json:"identityDescriptor,omitempty"`
	AllowPermissionBit *int   `json:"allowPermissionBit,omitempty"`
	DenyPermissionBit  *int   `json:"denyPermissionBit,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo,azuredevops}
//+kubebuilder:printcolumn:name="ALLOW",type="string",JSONPath=".status.allowPermissionBit"
//+kubebuilder:printcolumn:name="DENY",type="string",JSONPath=".status.denyPermissionBit"
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=10
//+kubebuilder:printcolumn:name="DESCRIPTOR",type="string",JSONPath=".status.identityDescriptor",priority=10

// RepositoryPermission is the Schema for the RepositoryPermissions API
type RepositoryPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositoryPermissionSpec   `json:"spec,omitempty"`
	Status RepositoryPermissionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepositoryPermissionList contains a list of RepositoryPermission
type RepositoryPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RepositoryPermission `json:"items"`
}
