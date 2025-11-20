# gardener-extension-example

The `gardener-extension-example` repo provides the skeleton for an example Gardener extension.

# Requirements

- [Go 1.25.x](https://go.dev/) or later
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://www.docker.com/) for local development
- [Gardener Local Setup](https://gardener.cloud/docs/gardener/local_setup/) for local development

# Code structure

The project repo uses the following code structure.

| Package           | Description                                                                              |
|-------------------|------------------------------------------------------------------------------------------|
| `cmd`             | Command-line application of the extension                                                |
| `pkg/apis`        | Extension API types, e.g. configuration spec, etc.                                       |
| `pkg/actuator`    | Implementations for the Gardener Extension Actuator interfaces                           |
| `pkg/controller`  | Utility wrappers for creating Kubernetes reconcilers for Gardener Actuators              |
| `pkg/healthcheck` | Utility wrappers for creating healthcheck reconcilers for Gardener extensions            |
| `pkg/heartbeat`   | Utility wrappers for creating heartbeat reconcilers for Gardener extensions              |
| `pkg/metrics`     | Metrics emitted by the extension                                                         |
| `pkg/mgr`         | Utility wrappers for creating `controller-runtime` managers using functional options API |
| `pkg/version`     | Version metadata information about the extension                                         |
| `internal/tools`  | Go-based tools used for testing and linting the project                                  |
| `charts`          | Helm charts for deploying the extension                                                  |
| `examples`        | Example Kubernetes resources, which can be used in a dev environment                     |
| `test`            | Various files (e.g. schemas, CRDs, etc.), used during testing                            |

# Usage

You can enable the extension for a [Gardener Shoot
cluster](https://gardener.cloud/docs/glossary/_index#gardener-glossary) by
updating the `.spec.extensions` of your shoot manifest.

``` yaml
...

spec:
  extensions:
    - type: example
      providerConfig:
        apiVersion: example.extensions.gardener.cloud/v1alpha1
        kind: ExampleConfig
        spec:
          foo: bar
```

# Development

For local development of the `gardener-extension-example` it is recommended that
you setup a development [Gardener](https://gardener.cloud/docs/gardener/local_setup/)
environment.

For more details on how to do that, please check the following documents.

- [Gardener: Local setup requirements](https://gardener.cloud/docs/gardener/local_setup/)
- [Gardener: Getting Started Locally](https://gardener.cloud/docs/gardener/deployment/getting_started_locally/)

Before you continue with the next steps, make sure that you configure your
`KUBECONFIG` to point to the kubeconfig file created by Gardener for you. This
file will be located in the
`/path/to/gardener/example/gardener-local/kind/local/kubeconfig` path after
creating a new local dev shoot cluster.

``` shell
export KUBECONFIG=/path/to/gardener/example/gardener-local/kind/local/kubeconfig
```

In order to build a binary of the extension, you can use the following command.

``` shell
make build
```

The resulting binary can be found in `bin/extension`.

In order to build a Docker image of the extension, you can use the following
command.

``` shell
make docker-build
```

You can use the following command in order to load the OCI image to the nodes of
your local Gardener cluster, which is running in
[kind](https://kind.sigs.k8s.io/).

``` shell
make kind-load-image
```

The Helm charts, which are used by the
[gardenlet](https://gardener.cloud/docs/gardener/concepts/gardenlet/) for
deploying the extension can be pushed to the local OCI registry using the
following command.

``` shell
make helm-load-chart
```

In the [./examples/dev-setup](./examples/dev-setup) directory you can find
[kustomize](https://kustomize.io/]) resources, which can be used to create the
`ControllerDeployment` and `ControllerRegistration` resources.

For more information about `ControllerDeployment` and `ControllerRegistration`
resources, please make sure to check the [Registering Extension
Controllers](https://gardener.cloud/docs/gardener/extensions/registration/)
documentation.

The `deploy` target takes care of deploying your extension in a local Gardener
environment. It does the following.

1. Builds a Docker image of the extension
2. Loads the image into the `kind` cluster nodes
3. Packages the Helm charts and pushes them to the local registry
4. Deploys the `ControllerDeployment` and `ControllerRegistration` resources

``` shell
make deploy
```

Verify that we have successfully created the `ControllerDeployment` and
`ControllerRegistration` resources.

``` shell
$ kubectl get controllerregistrations,controllerdeployments gardener-extension-example
NAME                                                                    RESOURCES           AGE
controllerregistration.core.gardener.cloud/gardener-extension-example   Extension/example   40s

NAME                                                                  AGE
controllerdeployment.core.gardener.cloud/gardener-extension-example   40s
```

Finally, we can create an example shoot with our extension enabled. The
[examples/shoot.yaml](./examples/shoot.yaml) file provides a ready-to-use shoot
manifest with the extension enabled and configured.

``` shell
kubectl apply -f examples/shoot.yaml
```

Once we create the shoot cluster, `gardenlet` will start deploying our
`gardener-extension-example`, since it is required by our shoot. Verify that the
extension has been successfully installed by checking the corresponding
`ControllerInstallation` resource.

``` shell
$ kubectl get controllerinstallations.core.gardener.cloud
NAME                               REGISTRATION                 SEED    VALID   INSTALLED   HEALTHY   PROGRESSING   AGE
gardener-extension-example-tktwt   gardener-extension-example   local   True    True        True      False         103s
```

After your shoot cluster has been successfully created and reconciled, verify
that the extension is healthy.

``` shell
$ kubectl --namespace shoot--local--local get extensions
NAME      TYPE      STATUS      AGE
example   example   Succeeded   85m
```

In order to trigger reconciliation of the extension you can annotate the
extension resource.

``` shell
kubectl --namespace shoot--local--local annotate extensions example gardener.cloud/operation=reconcile
```

# Tests

In order to run the tests use the command below:

``` shell
make test
```

In order to test the Helm chart and the manifests provided by it you can run the
following command.

``` shell
make check-helm
```

In order to test the example resources from the `examples/` directory you can
run the following command.

``` shell
make check-examples
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

# Contributing

`gardener-extension-example` is hosted on
[Github](https://github.com/dnaeon/gardener-extension-example).

Please contribute by reporting issues, suggesting features or by sending patches
using pull requests.

# License

This project is Open Source and licensed under [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0).
