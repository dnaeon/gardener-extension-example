// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gardener-extension-example/pkg/apis/config"
	"gardener-extension-example/pkg/apis/config/validation"
)

var _ = Describe("Validation Tests", Ordered, func() {
	It("should detect invalid config", func() {
		cfg := config.ExampleConfig{}
		err := validation.Validate(cfg)
		Expect(err).Should(HaveOccurred())
	})

	It("should successfully validate correct config", func() {
		cfg := config.ExampleConfig{
			Spec: config.ExampleConfigSpec{
				Foo: "bar",
			},
		}
		err := validation.Validate(cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	// TODO(user): additional tests
})
