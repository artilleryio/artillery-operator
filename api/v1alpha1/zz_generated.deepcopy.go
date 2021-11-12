//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *External) DeepCopyInto(out *External) {
	*out = *in
	in.Payload.DeepCopyInto(&out.Payload)
	in.Processor.DeepCopyInto(&out.Processor)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new External.
func (in *External) DeepCopy() *External {
	if in == nil {
		return nil
	}
	out := new(External)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadTest) DeepCopyInto(out *LoadTest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadTest.
func (in *LoadTest) DeepCopy() *LoadTest {
	if in == nil {
		return nil
	}
	out := new(LoadTest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LoadTest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadTestList) DeepCopyInto(out *LoadTestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LoadTest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadTestList.
func (in *LoadTestList) DeepCopy() *LoadTestList {
	if in == nil {
		return nil
	}
	out := new(LoadTestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LoadTestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadTestSpec) DeepCopyInto(out *LoadTestSpec) {
	*out = *in
	in.TestScript.DeepCopyInto(&out.TestScript)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadTestSpec.
func (in *LoadTestSpec) DeepCopy() *LoadTestSpec {
	if in == nil {
		return nil
	}
	out := new(LoadTestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadTestStatus) DeepCopyInto(out *LoadTestStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadTestStatus.
func (in *LoadTestStatus) DeepCopy() *LoadTestStatus {
	if in == nil {
		return nil
	}
	out := new(LoadTestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Main) DeepCopyInto(out *Main) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Main.
func (in *Main) DeepCopy() *Main {
	if in == nil {
		return nil
	}
	out := new(Main)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Payload) DeepCopyInto(out *Payload) {
	*out = *in
	if in.ConfigMaps != nil {
		in, out := &in.ConfigMaps, &out.ConfigMaps
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Payload.
func (in *Payload) DeepCopy() *Payload {
	if in == nil {
		return nil
	}
	out := new(Payload)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Processor) DeepCopyInto(out *Processor) {
	*out = *in
	out.Main = in.Main
	in.Related.DeepCopyInto(&out.Related)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Processor.
func (in *Processor) DeepCopy() *Processor {
	if in == nil {
		return nil
	}
	out := new(Processor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Related) DeepCopyInto(out *Related) {
	*out = *in
	if in.ConfigMaps != nil {
		in, out := &in.ConfigMaps, &out.ConfigMaps
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Related.
func (in *Related) DeepCopy() *Related {
	if in == nil {
		return nil
	}
	out := new(Related)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TestScript) DeepCopyInto(out *TestScript) {
	*out = *in
	out.Config = in.Config
	in.External.DeepCopyInto(&out.External)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TestScript.
func (in *TestScript) DeepCopy() *TestScript {
	if in == nil {
		return nil
	}
	out := new(TestScript)
	in.DeepCopyInto(out)
	return out
}
