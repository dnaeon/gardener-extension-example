// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExampleConfigSpec defines the desired state of [ExampleConfig]
type ExampleConfigSpec struct {
	// Foo is foo
	Foo string `json:"foo,omitzero"`

	// TODO(user): insert additional spec fields
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExampleConfig is the schema for the exampleconfigs API
// +k8s:openapi-gen=true
type ExampleConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// Spec provides the extension configuration spec.
	Spec ExampleConfigSpec `json:"spec,omitzero"`
}
