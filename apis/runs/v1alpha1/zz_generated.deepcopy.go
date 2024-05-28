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
func (in *BuildResourceParameters) DeepCopyInto(out *BuildResourceParameters) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildResourceParameters.
func (in *BuildResourceParameters) DeepCopy() *BuildResourceParameters {
	if in == nil {
		return nil
	}
	out := new(BuildResourceParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerResourceParameters) DeepCopyInto(out *ContainerResourceParameters) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerResourceParameters.
func (in *ContainerResourceParameters) DeepCopy() *ContainerResourceParameters {
	if in == nil {
		return nil
	}
	out := new(ContainerResourceParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PackageResourceParameters) DeepCopyInto(out *PackageResourceParameters) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PackageResourceParameters.
func (in *PackageResourceParameters) DeepCopy() *PackageResourceParameters {
	if in == nil {
		return nil
	}
	out := new(PackageResourceParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PipelineResourceParameters) DeepCopyInto(out *PipelineResourceParameters) {
	*out = *in
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PipelineResourceParameters.
func (in *PipelineResourceParameters) DeepCopy() *PipelineResourceParameters {
	if in == nil {
		return nil
	}
	out := new(PipelineResourceParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryResourceParameters) DeepCopyInto(out *RepositoryResourceParameters) {
	*out = *in
	if in.RefName != nil {
		in, out := &in.RefName, &out.RefName
		*out = new(string)
		**out = **in
	}
	if in.Token != nil {
		in, out := &in.Token, &out.Token
		*out = new(string)
		**out = **in
	}
	if in.TokenType != nil {
		in, out := &in.TokenType, &out.TokenType
		*out = new(string)
		**out = **in
	}
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryResourceParameters.
func (in *RepositoryResourceParameters) DeepCopy() *RepositoryResourceParameters {
	if in == nil {
		return nil
	}
	out := new(RepositoryResourceParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Run) DeepCopyInto(out *Run) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Run.
func (in *Run) DeepCopy() *Run {
	if in == nil {
		return nil
	}
	out := new(Run)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Run) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunList) DeepCopyInto(out *RunList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Run, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunList.
func (in *RunList) DeepCopy() *RunList {
	if in == nil {
		return nil
	}
	out := new(RunList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RunList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunPipelineParameters) DeepCopyInto(out *RunPipelineParameters) {
	*out = *in
	if in.PreviewRun != nil {
		in, out := &in.PreviewRun, &out.PreviewRun
		*out = new(bool)
		**out = **in
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(RunResourcesParameters)
		(*in).DeepCopyInto(*out)
	}
	if in.StagesToSkip != nil {
		in, out := &in.StagesToSkip, &out.StagesToSkip
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TemplateParameters != nil {
		in, out := &in.TemplateParameters, &out.TemplateParameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Variables != nil {
		in, out := &in.Variables, &out.Variables
		*out = make(map[string]Variable, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.YamlOverride != nil {
		in, out := &in.YamlOverride, &out.YamlOverride
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunPipelineParameters.
func (in *RunPipelineParameters) DeepCopy() *RunPipelineParameters {
	if in == nil {
		return nil
	}
	out := new(RunPipelineParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunResourcesParameters) DeepCopyInto(out *RunResourcesParameters) {
	*out = *in
	if in.Builds != nil {
		in, out := &in.Builds, &out.Builds
		*out = make(map[string]BuildResourceParameters, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make(map[string]ContainerResourceParameters, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Packages != nil {
		in, out := &in.Packages, &out.Packages
		*out = make(map[string]PackageResourceParameters, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Pipelines != nil {
		in, out := &in.Pipelines, &out.Pipelines
		*out = make(map[string]PipelineResourceParameters, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Repositories != nil {
		in, out := &in.Repositories, &out.Repositories
		*out = make(map[string]RepositoryResourceParameters, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunResourcesParameters.
func (in *RunResourcesParameters) DeepCopy() *RunResourcesParameters {
	if in == nil {
		return nil
	}
	out := new(RunResourcesParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunSpec) DeepCopyInto(out *RunSpec) {
	*out = *in
	out.ManagedSpec = in.ManagedSpec
	if in.RunParameters != nil {
		in, out := &in.RunParameters, &out.RunParameters
		*out = new(RunPipelineParameters)
		(*in).DeepCopyInto(*out)
	}
	if in.PipelineRef != nil {
		in, out := &in.PipelineRef, &out.PipelineRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.ConnectorConfigRef != nil {
		in, out := &in.ConnectorConfigRef, &out.ConnectorConfigRef
		*out = new(v1.Reference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunSpec.
func (in *RunSpec) DeepCopy() *RunSpec {
	if in == nil {
		return nil
	}
	out := new(RunSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunStatus) DeepCopyInto(out *RunStatus) {
	*out = *in
	in.ManagedStatus.DeepCopyInto(&out.ManagedStatus)
	if in.Id != nil {
		in, out := &in.Id, &out.Id
		*out = new(int)
		**out = **in
	}
	if in.PipelineId != nil {
		in, out := &in.PipelineId, &out.PipelineId
		*out = new(int)
		**out = **in
	}
	if in.State != nil {
		in, out := &in.State, &out.State
		*out = new(string)
		**out = **in
	}
	if in.Url != nil {
		in, out := &in.Url, &out.Url
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunStatus.
func (in *RunStatus) DeepCopy() *RunStatus {
	if in == nil {
		return nil
	}
	out := new(RunStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Variable) DeepCopyInto(out *Variable) {
	*out = *in
	if in.IsSecret != nil {
		in, out := &in.IsSecret, &out.IsSecret
		*out = new(bool)
		**out = **in
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Variable.
func (in *Variable) DeepCopy() *Variable {
	if in == nil {
		return nil
	}
	out := new(Variable)
	in.DeepCopyInto(out)
	return out
}
