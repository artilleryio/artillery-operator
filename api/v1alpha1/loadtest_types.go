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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Payload struct {
	ConfigMaps []string `json:"configMaps,omitempty"`
}

type Main struct {
	ConfigMap string `json:"configMap,omitempty"`
}
type Related struct {
	ConfigMaps []string `json:"configMaps,omitempty"`
}

type Processor struct {
	Main    Main    `json:"main,omitempty"`
	Related Related `json:"related,omitempty"`
}

type External struct {
	Payload   Payload   `json:"payload,omitempty"`
	Processor Processor `json:"processor,omitempty"`
}

type Config struct {
	ConfigMap string `json:"configMap,omitempty"`
}

type TestScript struct {
	Config   Config   `json:"config,omitempty"`
	External External `json:"external,omitempty"`
}

// LoadTestSpec defines the desired state of LoadTest
type LoadTestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Count       int        `json:"count,omitempty"`
	Environment string     `json:"environment,omitempty"`
	TestScript  TestScript `json:"testScript,omitempty"`
}

type Workers struct {
	Running []string `json:"running,omitempty"`
	Pending []string `json:"pending,omitempty"`
}

// LoadTestStatus defines the observed state of LoadTest
type LoadTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Workers Workers `json:"workers,omitempty"`
	Started bool    `json:"started,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// LoadTest is the Schema for the loadtests API
type LoadTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadTestSpec   `json:"spec,omitempty"`
	Status LoadTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadTestList contains a list of LoadTest
type LoadTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoadTest{}, &LoadTestList{})
}