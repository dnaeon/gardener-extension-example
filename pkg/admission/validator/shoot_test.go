// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator_test

import (
	"context"
	"encoding/json"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/ptr"

	exampleactuator "gardener-extension-example/pkg/actuator/example"
	"gardener-extension-example/pkg/admission/validator"
	"gardener-extension-example/pkg/apis/config"
)

var _ = Describe("Shoot Validator", Ordered, func() {
	var (
		ctx                = context.TODO()
		providerConfigData []byte
		decoder            = serializer.NewCodecFactory(scheme.Scheme, serializer.EnableStrict).UniversalDecoder()
		shootValidator     extensionswebhook.Validator
		shoot              *core.Shoot
		projectNamespace   = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "garden-local",
			},
		}
		providerConfig = config.ExampleConfig{
			Spec: config.ExampleConfigSpec{
				Foo: "bar",
			},
		}
	)

	BeforeAll(func() {
		var err error
		providerConfigData, err = json.Marshal(providerConfig)
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		var err error
		shootValidator, err = validator.NewShootValidator(decoder)
		Expect(err).NotTo(HaveOccurred())
		shoot = &core.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "local",
				Namespace: projectNamespace.Name,
			},
			Spec: core.ShootSpec{
				SeedName: ptr.To("local"),
				Provider: core.Provider{
					Type: "local",
				},
				Region: "local",
			},
		}
	})

	It("should successfully validate provider config", func() {
		// Ensure we have the extension enabled with proper provider config
		shoot.Spec.Extensions = []core.Extension{
			{
				Type: exampleactuator.ExtensionType,
				ProviderConfig: &runtime.RawExtension{
					Raw: providerConfigData,
				},
			},
		}

		Expect(shootValidator.Validate(ctx, shoot, nil)).NotTo(HaveOccurred())
	})

	It("should fail to create shoot validator with invalid decoder", func() {
		_, err := validator.NewShootValidator(nil)
		Expect(err).To(MatchError(ContainSubstring("invalid decoder specified")))
	})

	It("should fail to validate when extension is not defined", func() {
		err := shootValidator.Validate(ctx, shoot, nil)
		Expect(err).To(MatchError(ContainSubstring("not found in shoot spec")))
	})

	It("should fail to validate when extension provider config is not defined", func() {
		// Ensure we have the extension enabled with proper provider config
		shoot.Spec.Extensions = []core.Extension{
			{
				Type: exampleactuator.ExtensionType,
			},
		}
		err := shootValidator.Validate(ctx, shoot, nil)
		Expect(err).To(MatchError(ContainSubstring("no provider config specified")))
	})

	// TODO(user): additional tests
})
