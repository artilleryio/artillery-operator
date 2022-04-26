/*
 * Copyright (c) 2021-2022.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with
 * this file, You can obtain one at
 *
 *   http://mozilla.org/MPL/2.0/
 */

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Payload   *Payload   `json:"payload,omitempty"`
	Processor *Processor `json:"processor,omitempty"`
}

type Config struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	ConfigMap string `json:"configMap"`
}

type TestScript struct {
	// +kubebuilder:validation:Required
	Config   Config    `json:"config"`
	External *External `json:"external,omitempty"`
}

// LoadTestSpec defines the desired state of LoadTest
type LoadTestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Count       int    `json:"count,omitempty"`
	Environment string `json:"environment,omitempty"`

	// +kubebuilder:validation:Required
	TestScript TestScript `json:"testScript"`
}

// LoadTestConditionType creates types for K8s Conditions created by the operator
type LoadTestConditionType string

// These are valid conditions of a load-test.
const (
	// LoadTestProgressing means the load test's workers are executing tests against a test script target.
	LoadTestProgressing LoadTestConditionType = "Progressing"
	// LoadTestCompleted means the load test has completed its execution.
	LoadTestCompleted LoadTestConditionType = "Completed"
)

// LoadTestCondition provides a standard mechanism for higher-level status reporting
type LoadTestCondition struct {
	// Type of job condition, Progressing, Complete or Failed.
	Type LoadTestConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status"`
	// Last time the condition was checked.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// LoadTestStatus defines the observed state of LoadTest.
type LoadTestStatus struct {
	// Important: Run "make" to regenerate code after modifying this file.
	Conditions []LoadTestCondition `json:"conditions,omitempty"`

	// Represents time when the loadtest controller started processing a loadtest.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the loadtest was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// The completion time is only set when the loadtest finishes successfully.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Formatted duration of time required to complete the load test.
	// This is calculated using StartTime and CompletionTime
	Duration string `json:"duration,omitempty"`

	// The number of actively running LoadTest worker pods.
	Active int32 `json:"active,omitempty"`

	// The number of LoadTest worker pods which reached phase Succeeded.
	Succeeded int32 `json:"succeeded,omitempty"`

	// The number of LoadTest worker pods which reached phase Failed.
	Failed int32 `json:"failed,omitempty"`

	// Formatted load test worker pod completions calculated from the underlying succeeded jobs vs configured
	// job completions/parallelism.
	Completions string `json:"completions,omitempty"`

	// The image used to run the load tests.
	Image string `json:"image,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Completions",type="string",JSONPath=`.status.completions`
// +kubebuilder:printcolumn:name="Duration",type="string",JSONPath=`.status.duration`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.spec.environment`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.status.image`,priority=10

// LoadTest is the Schema for the loadTests API.
type LoadTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadTestSpec   `json:"spec,omitempty"`
	Status LoadTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadTestList contains a list of LoadTest.
type LoadTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoadTest{}, &LoadTestList{})
}
