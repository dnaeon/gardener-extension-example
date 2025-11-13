// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// TODO: command to generate the controller-registration manifest

package main

import (
	"os"

	"github.com/urfave/cli/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	managercmd "gardener-extension-example/cmd/extension/internal/manager"
	"gardener-extension-example/pkg/version"
)

func main() {
	app := &cli.Command{
		Name:                  "gardener-extension-example",
		Version:               version.Version,
		EnableShellCompletion: true,
		Usage:                 "an example gardener extension",
		Commands: []*cli.Command{
			managercmd.New(),
		},
	}

	ctx := ctrl.SetupSignalHandler()
	if err := app.Run(ctx, os.Args); err != nil {
		ctrllog.Log.Error(err, "failed to start extension")
		os.Exit(1)
	}
}
