// +kubebuilder:object:generate=true
// +groupName=azuredevops.krateo.io
// +versionName=v1alpha2
package v1alpha2

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "azuredevops.krateo.io"
	Version = "v1alpha2"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

var (
	PipelinePermissionKind             = reflect.TypeOf(PipelinePermission{}).Name()
	PipelinePermissionGroupKind        = schema.GroupKind{Group: Group, Kind: PipelinePermissionKind}.String()
	PipelinePermissionKindAPIVersion   = PipelinePermissionKind + "." + SchemeGroupVersion.String()
	PipelinePermissionGroupVersionKind = SchemeGroupVersion.WithKind(PipelinePermissionKind)
)

func init() {
	SchemeBuilder.Register(&PipelinePermission{}, &PipelinePermissionList{})
}
