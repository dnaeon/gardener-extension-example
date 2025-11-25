// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExampleConfigSpec defines the desired state of [ExampleConfig]
type ExampleConfigSpec struct {
	// Foo is foo
	Foo string

	// TODO(user): insert additional spec fields
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExampleConfig is the schema for the exampleconfigs API
type ExampleConfig struct {
	metav1.TypeMeta

	// Spec provides the extension configuration spec.
	Spec ExampleConfigSpec
}
