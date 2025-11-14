// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/extensions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		cluster = &extensions.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: shootNamespace.Name,
			},
			CloudProfile: &corev1beta1.CloudProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name: "local",
				},
				Spec: corev1beta1.CloudProfileSpec{
					Type: "local",
				},
			},
			Seed: &corev1beta1.Seed{
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
			},
			Shoot: &corev1beta1.Shoot{
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
		Expect(act.ExtensionClasses()).To(Equal([]extensionsv1alpha1.ExtensionClass{extensionsv1alpha1.ExtensionClassShoot}))
	})

	// TODO: add test with and without cluster
	_ = cluster

	Context("When reconciling an extension resource", func() {
		It("should successfully reconcile the resource", func() {
			act, err := actuator.New(actuatorOpts...)
			Expect(err).NotTo(HaveOccurred())
			Expect(act).NotTo(BeNil())
			Expect(act.Reconcile(ctx, logger, extResource)).To(Succeed())

			// TODO(user): Add more tests covering the various scenarios
		})
	})
})
