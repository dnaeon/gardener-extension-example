// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"context"
	"errors"
	"fmt"
	"slices"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	"github.com/gardener/gardener/pkg/apis/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"gardener-extension-example/pkg/actuator"
	"gardener-extension-example/pkg/apis/config"
	"gardener-extension-example/pkg/apis/config/validation"
)

// ErrInvalidShoot is an error, which is returned when an invalid shoot resource
// was provided.
var ErrInvalidShoot = errors.New("invalid shoot resource provided")

// ErrExtensionNotFound is an error, which is returned when the extension spec
// was not found.
var ErrExtensionNotFound = errors.New("extension not found")

// ErrInvalidExtensionConfig is an error, which is returned when the extension
// configuration is found to be invalid.
var ErrInvalidExtensionConfig = errors.New("invalid extension config")

// ErrInvalidValidator is an error which is returned when creating an
// [extensionswebhook.Validator] with invalid settings.
var ErrInvalidValidator = errors.New("invalid validator")

// shootValidator is an implementation of [extensionswebhook.Validator], which
// validates the provider configuration of the extension from a [core.Shoot]
// spec.
type shootValidator struct {
	reader        client.Reader
	decoder       runtime.Decoder
	extensionType string
}

var _ extensionswebhook.Validator = &shootValidator{}

// newShootValidator returns a new [extensionswebhook.Validator], which validates
// the extension provider config.
func newShootValidator(reader client.Reader, decoder runtime.Decoder) (*shootValidator, error) {
	validator := &shootValidator{
		reader:        reader,
		decoder:       decoder,
		extensionType: actuator.ExtensionType,
	}

	if reader == nil {
		return nil, fmt.Errorf("%w: no reader specified", ErrInvalidValidator)
	}

	if decoder == nil {
		return nil, fmt.Errorf("%w: no decoder specified", ErrInvalidValidator)
	}

	return validator, nil
}

// Validate implements the [extensionswebhook.Validator] interface. This method
// validates the extension provider configuration from a [core.Shoot] spec.
func (v *shootValidator) Validate(ctx context.Context, newObj, oldObj client.Object) error {
	newShoot, ok := newObj.(*core.Shoot)
	if !ok {
		return fmt.Errorf("invalid object type: %T", newObj)
	}
	oldShoot, ok := oldObj.(*core.Shoot)
	if !ok {
		oldShoot = nil
	}

	return v.validateExtension(newShoot, oldShoot)
}

// getExtension returns the [core.Extension] by extracting it from the given
// [core.Shoot] object.
func (v *shootValidator) getExtension(obj *core.Shoot) (core.Extension, error) {
	if obj == nil {
		return core.Extension{}, ErrInvalidShoot
	}

	idx := slices.IndexFunc(obj.Spec.Extensions, func(ext core.Extension) bool {
		return ext.Type == v.extensionType
	})

	if idx == -1 {
		return core.Extension{}, ErrExtensionNotFound
	}

	return obj.Spec.Extensions[idx], nil
}

// validateExtension validates the extension configuration from the given
// [core.Shoot] specs.
func (v *shootValidator) validateExtension(newObj *core.Shoot, _ *core.Shoot) error {
	ext, err := v.getExtension(newObj)
	if err != nil {
		return err
	}

	// Extension is disabled, nothing to validate
	if ext.Disabled != nil && *ext.Disabled {
		return nil
	}

	if ext.ProviderConfig == nil {
		return fmt.Errorf("%w: no provider config specified", ErrInvalidExtensionConfig)
	}

	var cfg config.ExampleConfig
	if err := runtime.DecodeInto(v.decoder, ext.ProviderConfig.Raw, &cfg); err != nil {
		return fmt.Errorf("%w: invalid provider spec configuration: %w", ErrInvalidExtensionConfig, err)
	}

	if err := validation.Validate(cfg); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidExtensionConfig, err.Error())
	}

	return nil
}

// NewShootValidatorWebhook returns a new [extensionswebhook.Webhook], which
// validates extension configuration defined in a [core.Shoot] object.
func NewShootValidatorWebhook(mgr manager.Manager) (*extensionswebhook.Webhook, error) {
	decoder := serializer.NewCodecFactory(mgr.GetScheme(), serializer.EnableStrict).UniversalDecoder()
	validator, err := newShootValidator(mgr.GetAPIReader(), decoder)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("validator.%s", validator.extensionType)
	extensionLabel := fmt.Sprintf("extensions.extensions.gardener.cloud/%s", validator.extensionType)
	path := "/webhooks/validate"

	args := extensionswebhook.Args{
		Provider: validator.extensionType,
		Name:     name,
		Path:     path,
		Validators: map[extensionswebhook.Validator][]extensionswebhook.Type{
			validator: {{Obj: &core.Shoot{}}},
		},
		Target: extensionswebhook.TargetSeed,
		ObjectSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				extensionLabel: "true",
			},
		},
	}

	return extensionswebhook.New(mgr, args)
}
