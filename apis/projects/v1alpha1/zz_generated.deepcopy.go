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
func (in *Capabilities) DeepCopyInto(out *Capabilities) {
	*out = *in
	if in.Versioncontrol != nil {
		in, out := &in.Versioncontrol, &out.Versioncontrol
		*out = new(Versioncontrol)
		**out = **in
	}
	if in.ProcessTemplate != nil {
		in, out := &in.ProcessTemplate, &out.ProcessTemplate
		*out = new(ProcessTemplate)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Capabilities.
func (in *Capabilities) DeepCopy() *Capabilities {
	if in == nil {
		return nil
	}
	out := new(Capabilities)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProcessTemplate) DeepCopyInto(out *ProcessTemplate) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProcessTemplate.
func (in *ProcessTemplate) DeepCopy() *ProcessTemplate {
	if in == nil {
		return nil
	}
	out := new(ProcessTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TeamProject) DeepCopyInto(out *TeamProject) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TeamProject.
func (in *TeamProject) DeepCopy() *TeamProject {
	if in == nil {
		return nil
	}
	out := new(TeamProject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TeamProject) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TeamProjectList) DeepCopyInto(out *TeamProjectList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TeamProject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TeamProjectList.
func (in *TeamProjectList) DeepCopy() *TeamProjectList {
	if in == nil {
		return nil
	}
	out := new(TeamProjectList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TeamProjectList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TeamProjectSpec) DeepCopyInto(out *TeamProjectSpec) {
	*out = *in
	out.ManagedSpec = in.ManagedSpec
	if in.ConnectorConfigRef != nil {
		in, out := &in.ConnectorConfigRef, &out.ConnectorConfigRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.Visibility != nil {
		in, out := &in.Visibility, &out.Visibility
		*out = new(string)
		**out = **in
	}
	in.Capabilities.DeepCopyInto(&out.Capabilities)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TeamProjectSpec.
func (in *TeamProjectSpec) DeepCopy() *TeamProjectSpec {
	if in == nil {
		return nil
	}
	out := new(TeamProjectSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TeamProjectStatus) DeepCopyInto(out *TeamProjectStatus) {
	*out = *in
	in.ManagedStatus.DeepCopyInto(&out.ManagedStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TeamProjectStatus.
func (in *TeamProjectStatus) DeepCopy() *TeamProjectStatus {
	if in == nil {
		return nil
	}
	out := new(TeamProjectStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Versioncontrol) DeepCopyInto(out *Versioncontrol) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Versioncontrol.
func (in *Versioncontrol) DeepCopy() *Versioncontrol {
	if in == nil {
		return nil
	}
	out := new(Versioncontrol)
	in.DeepCopyInto(out)
	return out
}
