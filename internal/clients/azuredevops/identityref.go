package azuredevops

type IdentityRef struct {
	// This field contains zero or more interesting links about the graph subject. These links may be invoked to obtain additional relationships or more detailed information about this graph subject.
	Links interface{} `json:"_links,omitempty"`
	// The descriptor is the primary way to reference the graph subject while the system is running. This field will uniquely identify the same graph subject across both Accounts and Organizations.
	Descriptor *string `json:"descriptor,omitempty"`
	// This is the non-unique display name of the graph subject. To change this field, you must alter its value in the source provider.
	DisplayName *string `json:"displayName,omitempty"`
	// This url is the full route to the source resource of this graph subject.
	Url *string `json:"url,omitempty"`
	// Deprecated - Can be retrieved by querying the Graph user referenced in the "self" entry of the IdentityRef "_links" dictionary
	DirectoryAlias *string `json:"directoryAlias,omitempty"`
	Id             *string `json:"id,omitempty"`
	// Deprecated - Available in the "avatar" entry of the IdentityRef "_links" dictionary
	ImageUrl *string `json:"imageUrl,omitempty"`
	// Deprecated - Can be retrieved by querying the Graph membership state referenced in the "membershipState" entry of the GraphUser "_links" dictionary
	Inactive *bool `json:"inactive,omitempty"`
	// Deprecated - Can be inferred from the subject type of the descriptor (Descriptor.IsAadUserType/Descriptor.IsAadGroupType)
	IsAadIdentity *bool `json:"isAadIdentity,omitempty"`
	// Deprecated - Can be inferred from the subject type of the descriptor (Descriptor.IsGroupType)
	IsContainer       *bool `json:"isContainer,omitempty"`
	IsDeletedInOrigin *bool `json:"isDeletedInOrigin,omitempty"`
	// Deprecated - not in use in most preexisting implementations of ToIdentityRef
	ProfileUrl *string `json:"profileUrl,omitempty"`
	// Deprecated - use Domain+PrincipalName instead
	UniqueName *string `json:"uniqueName,omitempty"`
}
