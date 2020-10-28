// +build !ignore_autogenerated

/*
Copyright 2020 GramLabs, Inc.

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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AmazonWebServices) DeepCopyInto(out *AmazonWebServices) {
	*out = *in
	if in.Cost != nil {
		in, out := &in.Cost, &out.Cost
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AmazonWebServices.
func (in *AmazonWebServices) DeepCopy() *AmazonWebServices {
	if in == nil {
		return nil
	}
	out := new(AmazonWebServices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Application) DeepCopyInto(out *Application) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = new(Parameters)
		(*in).DeepCopyInto(*out)
	}
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(Ingress)
		**out = **in
	}
	if in.Scenarios != nil {
		in, out := &in.Scenarios, &out.Scenarios
		*out = make([]Scenario, len(*in))
		copy(*out, *in)
	}
	if in.Objectives != nil {
		in, out := &in.Objectives, &out.Objectives
		*out = make([]Objective, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.CloudProvider != nil {
		in, out := &in.CloudProvider, &out.CloudProvider
		*out = new(CloudProvider)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Application.
func (in *Application) DeepCopy() *Application {
	if in == nil {
		return nil
	}
	out := new(Application)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Application) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudProvider) DeepCopyInto(out *CloudProvider) {
	*out = *in
	if in.GenericCloudProvider != nil {
		in, out := &in.GenericCloudProvider, &out.GenericCloudProvider
		*out = new(GenericCloudProvider)
		(*in).DeepCopyInto(*out)
	}
	if in.GCP != nil {
		in, out := &in.GCP, &out.GCP
		*out = new(GoogleCloudPlatform)
		(*in).DeepCopyInto(*out)
	}
	if in.AWS != nil {
		in, out := &in.AWS, &out.AWS
		*out = new(AmazonWebServices)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudProvider.
func (in *CloudProvider) DeepCopy() *CloudProvider {
	if in == nil {
		return nil
	}
	out := new(CloudProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerResources) DeepCopyInto(out *ContainerResources) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerResources.
func (in *ContainerResources) DeepCopy() *ContainerResources {
	if in == nil {
		return nil
	}
	out := new(ContainerResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CostObjective) DeepCopyInto(out *CostObjective) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CostObjective.
func (in *CostObjective) DeepCopy() *CostObjective {
	if in == nil {
		return nil
	}
	out := new(CostObjective)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenericCloudProvider) DeepCopyInto(out *GenericCloudProvider) {
	*out = *in
	if in.Cost != nil {
		in, out := &in.Cost, &out.Cost
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenericCloudProvider.
func (in *GenericCloudProvider) DeepCopy() *GenericCloudProvider {
	if in == nil {
		return nil
	}
	out := new(GenericCloudProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GoogleCloudPlatform) DeepCopyInto(out *GoogleCloudPlatform) {
	*out = *in
	if in.Cost != nil {
		in, out := &in.Cost, &out.Cost
		*out = make(v1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GoogleCloudPlatform.
func (in *GoogleCloudPlatform) DeepCopy() *GoogleCloudPlatform {
	if in == nil {
		return nil
	}
	out := new(GoogleCloudPlatform)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Ingress) DeepCopyInto(out *Ingress) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Ingress.
func (in *Ingress) DeepCopy() *Ingress {
	if in == nil {
		return nil
	}
	out := new(Ingress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Objective) DeepCopyInto(out *Objective) {
	*out = *in
	if in.Cost != nil {
		in, out := &in.Cost, &out.Cost
		*out = new(CostObjective)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Objective.
func (in *Objective) DeepCopy() *Objective {
	if in == nil {
		return nil
	}
	out := new(Objective)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Parameters) DeepCopyInto(out *Parameters) {
	*out = *in
	if in.ContainerResources != nil {
		in, out := &in.ContainerResources, &out.ContainerResources
		*out = new(ContainerResources)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Parameters.
func (in *Parameters) DeepCopy() *Parameters {
	if in == nil {
		return nil
	}
	out := new(Parameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Scenario) DeepCopyInto(out *Scenario) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Scenario.
func (in *Scenario) DeepCopy() *Scenario {
	if in == nil {
		return nil
	}
	out := new(Scenario)
	in.DeepCopyInto(out)
	return out
}