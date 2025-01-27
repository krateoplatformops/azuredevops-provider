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
func (in *ApprovalSettings) DeepCopyInto(out *ApprovalSettings) {
	*out = *in
	if in.Approvers != nil {
		in, out := &in.Approvers, &out.Approvers
		*out = make([]Approver, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.BlockedApprovers != nil {
		in, out := &in.BlockedApprovers, &out.BlockedApprovers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApprovalSettings.
func (in *ApprovalSettings) DeepCopy() *ApprovalSettings {
	if in == nil {
		return nil
	}
	out := new(ApprovalSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Approver) DeepCopyInto(out *Approver) {
	*out = *in
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(string)
		**out = **in
	}
	if in.ApproverRef != nil {
		in, out := &in.ApproverRef, &out.ApproverRef
		*out = new(v1.Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Approver.
func (in *Approver) DeepCopy() *Approver {
	if in == nil {
		return nil
	}
	out := new(Approver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheckConfiguration) DeepCopyInto(out *CheckConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheckConfiguration.
func (in *CheckConfiguration) DeepCopy() *CheckConfiguration {
	if in == nil {
		return nil
	}
	out := new(CheckConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheckConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheckConfigurationList) DeepCopyInto(out *CheckConfigurationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CheckConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheckConfigurationList.
func (in *CheckConfigurationList) DeepCopy() *CheckConfigurationList {
	if in == nil {
		return nil
	}
	out := new(CheckConfigurationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheckConfigurationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheckConfigurationSpec) DeepCopyInto(out *CheckConfigurationSpec) {
	*out = *in
	out.ManagedSpec = in.ManagedSpec
	if in.ConnectorConfigRef != nil {
		in, out := &in.ConnectorConfigRef, &out.ConnectorConfigRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.ProjectRef != nil {
		in, out := &in.ProjectRef, &out.ProjectRef
		*out = new(v1.Reference)
		**out = **in
	}
	in.Resource.DeepCopyInto(&out.Resource)
	in.ApprovalSettings.DeepCopyInto(&out.ApprovalSettings)
	out.TaskCheckSettings = in.TaskCheckSettings
	if in.ExtendsCheckSettings != nil {
		in, out := &in.ExtendsCheckSettings, &out.ExtendsCheckSettings
		*out = make([]ExtendsCheckSettings, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheckConfigurationSpec.
func (in *CheckConfigurationSpec) DeepCopy() *CheckConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(CheckConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheckConfigurationStatus) DeepCopyInto(out *CheckConfigurationStatus) {
	*out = *in
	in.ManagedStatus.DeepCopyInto(&out.ManagedStatus)
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheckConfigurationStatus.
func (in *CheckConfigurationStatus) DeepCopy() *CheckConfigurationStatus {
	if in == nil {
		return nil
	}
	out := new(CheckConfigurationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefinitionRef) DeepCopyInto(out *DefinitionRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefinitionRef.
func (in *DefinitionRef) DeepCopy() *DefinitionRef {
	if in == nil {
		return nil
	}
	out := new(DefinitionRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtendsCheckSettings) DeepCopyInto(out *ExtendsCheckSettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtendsCheckSettings.
func (in *ExtendsCheckSettings) DeepCopy() *ExtendsCheckSettings {
	if in == nil {
		return nil
	}
	out := new(ExtendsCheckSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resource) DeepCopyInto(out *Resource) {
	*out = *in
	if in.ResourceRef != nil {
		in, out := &in.ResourceRef, &out.ResourceRef
		*out = new(v1.Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resource.
func (in *Resource) DeepCopy() *Resource {
	if in == nil {
		return nil
	}
	out := new(Resource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TaskCheckSettings) DeepCopyInto(out *TaskCheckSettings) {
	*out = *in
	out.DefinitionRef = in.DefinitionRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TaskCheckSettings.
func (in *TaskCheckSettings) DeepCopy() *TaskCheckSettings {
	if in == nil {
		return nil
	}
	out := new(TaskCheckSettings)
	in.DeepCopyInto(out)
	return out
}
