package actuator

import (
	"context"

	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
)

// Name specifies the name of the actuator
const Name = "example"

// Actuator is an implementation of [extension.Actuator].
// TODO: add client to the actuator
type Actuator struct{}

// Option is a function, which configures the [Actuator].
type Option func(a *Actuator) error

// New creates a new actuator
func New(opts ...Option) (extension.Actuator, error) {
	act := &Actuator{}
	for _, opt := range opts {
		if err := opt(act); err != nil {
			return nil, err
		}
	}

	return act, nil
}

// Reconcile reconciles the [extensionsv1alpha1.Extension] resource by taking
// care of any resources managed by the [Actuator].
func (a *Actuator) Reconcile(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// TODO: boilerplate
	// TODO: logging
	// TODO: metrics
	return nil
}

// Delete deletes any resources managed by the [Actuator].
func (a *Actuator) Delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// TODO: boilerplate
	// TODO: logging
	// TODO: metrics
	return nil
}

// ForceDelete signals the [Actuator] to delete any resources managed by it,
// because of a force-delete event of the shoot cluster.
func (a *Actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// TODO: boilerplate
	// TODO: logging
	// TODO: metrics
	return nil
}

// Restore restores the resources managed by the extension [Actuator].
func (a *Actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// TODO: boilerplate
	// TODO: logging
	// TODO: metrics
	return a.Reconcile(ctx, logger, ex)
}

// Migrate signals the [Actuator] to reconcile the resources managed by it,
// because of a shoot control-plane migration event.
func (a *Actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// TODO: boilerplate
	// TODO: logging
	// TODO: metrics
	return a.Reconcile(ctx, logger, ex)
}
