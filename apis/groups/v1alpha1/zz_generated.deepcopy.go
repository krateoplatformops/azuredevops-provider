//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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
func (in *GroupIdentifier) DeepCopyInto(out *GroupIdentifier) {
	*out = *in
	if in.GroupsName != nil {
		in, out := &in.GroupsName, &out.GroupsName
		*out = new(string)
		**out = **in
	}
	if in.OriginID != nil {
		in, out := &in.OriginID, &out.OriginID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupIdentifier.
func (in *GroupIdentifier) DeepCopy() *GroupIdentifier {
	if in == nil {
		return nil
	}
	out := new(GroupIdentifier)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Groups) DeepCopyInto(out *Groups) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Groups.
func (in *Groups) DeepCopy() *Groups {
	if in == nil {
		return nil
	}
	out := new(Groups)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Groups) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GroupsList) DeepCopyInto(out *GroupsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Groups, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupsList.
func (in *GroupsList) DeepCopy() *GroupsList {
	if in == nil {
		return nil
	}
	out := new(GroupsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GroupsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GroupsSpec) DeepCopyInto(out *GroupsSpec) {
	*out = *in
	out.ManagedSpec = in.ManagedSpec
	if in.ConnectorConfigRef != nil {
		in, out := &in.ConnectorConfigRef, &out.ConnectorConfigRef
		*out = new(v1.Reference)
		**out = **in
	}
	in.Membership.DeepCopyInto(&out.Membership)
	in.GroupIdentifier.DeepCopyInto(&out.GroupIdentifier)
	if in.GroupsRefs != nil {
		in, out := &in.GroupsRefs, &out.GroupsRefs
		*out = make([]v1.Reference, len(*in))
		copy(*out, *in)
	}
	if in.TeamsRefs != nil {
		in, out := &in.TeamsRefs, &out.TeamsRefs
		*out = make([]v1.Reference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupsSpec.
func (in *GroupsSpec) DeepCopy() *GroupsSpec {
	if in == nil {
		return nil
	}
	out := new(GroupsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GroupsStatus) DeepCopyInto(out *GroupsStatus) {
	*out = *in
	in.ManagedStatus.DeepCopyInto(&out.ManagedStatus)
	if in.Descriptor != nil {
		in, out := &in.Descriptor, &out.Descriptor
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupsStatus.
func (in *GroupsStatus) DeepCopy() *GroupsStatus {
	if in == nil {
		return nil
	}
	out := new(GroupsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Membership) DeepCopyInto(out *Membership) {
	*out = *in
	if in.Organization != nil {
		in, out := &in.Organization, &out.Organization
		*out = new(string)
		**out = **in
	}
	if in.ProjectRef != nil {
		in, out := &in.ProjectRef, &out.ProjectRef
		*out = new(v1.Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Membership.
func (in *Membership) DeepCopy() *Membership {
	if in == nil {
		return nil
	}
	out := new(Membership)
	in.DeepCopyInto(out)
	return out
}
