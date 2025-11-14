# gardener-extension-example

The `gardener-extension-example` repo provides the skeleton for an example
Gardener extension.

# Requirements

- Go 1.25.x or later
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/) for local development

# Code structure

The project repo uses the following code structure.

| Package           | Description                                                                              |
|-------------------|------------------------------------------------------------------------------------------|
| `cmd`             | Command-line application of the extension                                                |
| `pkg/actuator`    | Implementations for the Gardener Extension Actuator interfaces                           |
| `pkg/controller`  | Utility wrappers for creating Kubernetes reconcilers for Gardener Actuators              |
| `pkg/healthcheck` | Utility wrappers for creating healthcheck reconcilers for Gardener extensions            |
| `pkg/heartbeat`   | Utility wrappers for creating heartbeat reconcilers for Gardener extensions              |
| `pkg/metrics`     | Metrics emitted by the extension                                                         |
| `pkg/mgr`         | Utility wrappers for creating `controller-runtime` managers using functional options API |
| `pkg/version`     | Version metadata information about the extension                                         |
| `internal/tools`  | Go-based tools used for testing and linting the project                                  |

# Building

In order to build the extension locally, execute the following command:

``` shell
make build
```

The resulting binary can be found in `bin/extension`.

In order to build a Docker image of the extension, execute the following
command:

``` shell
make docker-image
```

# Deploy

TODO: add deployment examples

# Usage

TODO: add usage examples

# Development

# Tests

Tests are using
[envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest) for
running a local control-plane.

In order to run the tests use the command below:

``` shell
make test
```

# Documentation

Make sure to check the following documents for more information about Gardener
Extensions and the available extensions API.

- [Gardener: Extensibility Overview](https://gardener.cloud/docs/gardener/extensions/)
- [Gardener: Registering Extension Controllers](https://gardener.cloud/docs/gardener/extensions/registration/)
- [Gardener: Extension Resources](https://github.com/gardener/gardener/tree/master/docs/extensions/resources)
- [Gardener: Extensions API Contract](https://github.com/gardener/gardener/blob/master/docs/extensions/resources/extension.md)
- [Gardener: How to Set Up a Gardener Landscape](https://gardener.cloud/docs/gardener/deployment/setup_gardener/)
- [Gardener: Extension Packages (Go)](https://github.com/gardener/gardener/tree/master/extensions/pkg)

Additional documentation worth checking out:

- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime/tree/main)
- [Envtest Binaries Manager](https://github.com/kubernetes-sigs/controller-runtime/tree/main/tools/setup-envtest)

# Contributing

`gardener-extension-example` is hosted on
[Github](https://github.com/dnaeon/gardener-extension-example).

Please contribute by reporting issues, suggesting features or by sending patches
using pull requests.

# License

This project is Open Source and licensed under [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
