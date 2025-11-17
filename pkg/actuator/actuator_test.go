// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator_test

import (
	"encoding/json"

	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/component-base/featuregate"
	"k8s.io/utils/ptr"

	"gardener-extension-example/pkg/actuator"
)

var _ = Describe("Actuator", Ordered, func() {
	var (
		decoder      = serializer.NewCodecFactory(scheme.Scheme, serializer.EnableStrict).UniversalDecoder()
		featureGates = make(map[featuregate.Feature]bool)
		actuatorOpts []actuator.Option

		projectNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "garden-local",
			},
		}
		shootNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "shoot--local--local",
			},
		}
		extResource = &extensionsv1alpha1.Extension{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: shootNamespace.Name,
			},
			Spec: extensionsv1alpha1.ExtensionSpec{
				DefaultSpec: extensionsv1alpha1.DefaultSpec{
					Type:  actuator.ExtensionType,
					Class: ptr.To(extensionsv1alpha1.ExtensionClassShoot),
				},
			},
		}
		cloudProfile = &corev1beta1.CloudProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: "local",
			},
			Spec: corev1beta1.CloudProfileSpec{
				Type: "local",
			},
		}
		seed = &corev1beta1.Seed{
			ObjectMeta: metav1.ObjectMeta{
				Name: "local",
			},
			Spec: corev1beta1.SeedSpec{
				Ingress: &corev1beta1.Ingress{
					Domain: "ingress.local.seed.local.gardener.cloud",
				},
				Provider: corev1beta1.SeedProvider{
					Type:   "local",
					Region: "local",
					Zones:  []string{"0"},
				},
			},
		}
		shoot = &corev1beta1.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "local",
				Namespace: projectNamespace.Name,
			},
			Spec: corev1beta1.ShootSpec{
				SeedName: ptr.To("local"),
				Provider: corev1beta1.Provider{
					Type: "local",
				},
				Region: "local",
			},
		}
	)

	BeforeAll(func() {
		actuatorOpts = []actuator.Option{
			actuator.WithClient(k8sClient),
			actuator.WithReader(k8sClient),
			actuator.WithGardenerVersion("1.0.0"),
			actuator.WithDecoder(decoder),
			actuator.WithGardenletFeatures(featureGates),
		}
	})

	It("should successfully create an actuator", func() {
		act, err := actuator.New(actuatorOpts...)

		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Name()).To(Equal(actuator.Name))
		Expect(act.ExtensionType()).To(Equal(actuator.ExtensionType))
		Expect(act.FinalizerSuffix()).To(Equal(actuator.FinalizerSuffix))
		Expect(act.ExtensionClass()).To(Equal(extensionsv1alpha1.ExtensionClassShoot))
	})

	It("should fail to reconcile when no cluster exists", func() {
		act, err := actuator.New(actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Reconcile(ctx, logger, extResource)).Error().Should(HaveOccurred())
	})

	It("should succeed on Reconcile", func() {
		cpData, err := json.Marshal(cloudProfile)
		Expect(err).NotTo(HaveOccurred())
		seedData, err := json.Marshal(seed)
		Expect(err).NotTo(HaveOccurred())
		shootData, err := json.Marshal(shoot)
		Expect(err).NotTo(HaveOccurred())

		cluster := &extensionsv1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: shootNamespace.Name,
			},
			Spec: extensionsv1alpha1.ClusterSpec{
				CloudProfile: runtime.RawExtension{
					Raw: cpData,
				},
				Seed: runtime.RawExtension{
					Raw: seedData,
				},
				Shoot: runtime.RawExtension{
					Raw: shootData,
				},
			},
		}
		Expect(k8sClient.Create(ctx, shootNamespace)).To(Succeed())
		Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

		act, err := actuator.New(actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Reconcile(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on Delete", func() {
		act, err := actuator.New(actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Delete(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on ForceDelete", func() {
		act, err := actuator.New(actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.ForceDelete(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on Restore", func() {
		act, err := actuator.New(actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Restore(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})

	It("should succeed on Migrate", func() {
		act, err := actuator.New(actuatorOpts...)
		Expect(err).NotTo(HaveOccurred())
		Expect(act).NotTo(BeNil())
		Expect(act.Migrate(ctx, logger, extResource)).To(Succeed())

		// TODO(user): Add more tests
	})
})
