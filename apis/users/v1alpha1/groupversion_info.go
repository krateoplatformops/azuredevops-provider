// +kubebuilder:object:generate=true
// +groupName=azuredevops.krateo.io
// +versionName=v1alpha1
package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "azuredevops.krateo.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

var (
	UsersKind             = reflect.TypeOf(Users{}).Name()
	UsersGroupKind        = schema.GroupKind{Group: Group, Kind: UsersKind}.String()
	UsersKindAPIVersion   = UsersKind + "." + SchemeGroupVersion.String()
	UsersGroupVersionKind = SchemeGroupVersion.WithKind(UsersKind)
)

func init() {
	SchemeBuilder.Register(&Users{}, &UsersList{})
}