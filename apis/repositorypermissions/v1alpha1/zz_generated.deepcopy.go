//go:build !ignore_autogenerated

/*
Copyright 2023 Kiratech SPA.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/krateoplatformops/provider-runtime/apis/common/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Identity) DeepCopyInto(out *Identity) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(string)
		**out = **in
	}
	if in.ProjectRef != nil {
		in, out := &in.ProjectRef, &out.ProjectRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Identity.
func (in *Identity) DeepCopy() *Identity {
	if in == nil {
		return nil
	}
	out := new(Identity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Permissions) DeepCopyInto(out *Permissions) {
	*out = *in
	if in.Identity != nil {
		in, out := &in.Identity, &out.Identity
		*out = new(Identity)
		(*in).DeepCopyInto(*out)
	}
	if in.AllowList != nil {
		in, out := &in.AllowList, &out.AllowList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DenyList != nil {
		in, out := &in.DenyList, &out.DenyList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Permissions.
func (in *Permissions) DeepCopy() *Permissions {
	if in == nil {
		return nil
	}
	out := new(Permissions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryPermission) DeepCopyInto(out *RepositoryPermission) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryPermission.
func (in *RepositoryPermission) DeepCopy() *RepositoryPermission {
	if in == nil {
		return nil
	}
	out := new(RepositoryPermission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RepositoryPermission) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryPermissionList) DeepCopyInto(out *RepositoryPermissionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RepositoryPermission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryPermissionList.
func (in *RepositoryPermissionList) DeepCopy() *RepositoryPermissionList {
	if in == nil {
		return nil
	}
	out := new(RepositoryPermissionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RepositoryPermissionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryPermissionSpec) DeepCopyInto(out *RepositoryPermissionSpec) {
	*out = *in
	out.ManagedSpec = in.ManagedSpec
	if in.RepositoryRef != nil {
		in, out := &in.RepositoryRef, &out.RepositoryRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.Permissions != nil {
		in, out := &in.Permissions, &out.Permissions
		*out = new(Permissions)
		(*in).DeepCopyInto(*out)
	}
	if in.ConnectorConfigRef != nil {
		in, out := &in.ConnectorConfigRef, &out.ConnectorConfigRef
		*out = new(v1.Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryPermissionSpec.
func (in *RepositoryPermissionSpec) DeepCopy() *RepositoryPermissionSpec {
	if in == nil {
		return nil
	}
	out := new(RepositoryPermissionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryPermissionStatus) DeepCopyInto(out *RepositoryPermissionStatus) {
	*out = *in
	in.ManagedStatus.DeepCopyInto(&out.ManagedStatus)
	if in.AllowPermissionBit != nil {
		in, out := &in.AllowPermissionBit, &out.AllowPermissionBit
		*out = new(int)
		**out = **in
	}
	if in.DenyPermissionBit != nil {
		in, out := &in.DenyPermissionBit, &out.DenyPermissionBit
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryPermissionStatus.
func (in *RepositoryPermissionStatus) DeepCopy() *RepositoryPermissionStatus {
	if in == nil {
		return nil
	}
	out := new(RepositoryPermissionStatus)
	in.DeepCopyInto(out)
	return out
}
