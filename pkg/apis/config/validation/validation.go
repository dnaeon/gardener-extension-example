// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"gardener-extension-example/pkg/apis/config"
)

// Validate validates the given [config.ExampleConfig]
func Validate(cfg config.ExampleConfig) error {
	allErrs := make(field.ErrorList, 0)

	if cfg.Spec.Foo == "" {
		allErrs = append(
			allErrs,
			field.Required(field.NewPath("spec.foo"), "empty value specified"),
		)
	}

	// TODO(user): validate any other config setting

	return allErrs.ToAggregate()
}
