// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"context"
	"errors"
	"net/url"
	"os"
	"slices"

	extensionswebhook "github.com/gardener/gardener/extensions/pkg/webhook"
	gardenercorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/urfave/cli/v3"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	configinstall "gardener-extension-example/pkg/apis/config/install"
	"gardener-extension-example/pkg/mgr"
)

// flags stores the webhook flags as provided from the command-line
type flags struct {
	extensionName               string
	metricsBindAddr             string
	healthProbeBindAddr         string
	leaderElection              bool
	leaderElectionID            string
	leaderElectionNamespace     string
	kubeconfig                  string
	zapLogLevel                 string
	zapLogFormat                string
	pprofBindAddr               string
	clientConnQPS               float32
	clientConnBurst             int32
	webhookServerHost           string
	webhookServerPort           int
	webhookServerCertDir        string
	webhookServerCertName       string
	webhookServerKeyName        string
	webhookConfigNamespace      string
	webhookConfigMode           string
	webhookConfigURL            string
	webhookConfigServicePort    int
	webhookConfigOwnerNamespace string
	gardenerVersion             string
	selfHostedShootCluster      bool
}

// getManager creates a new [ctrl.Manager] based on the parsed [flags].
func (f *flags) getManager(ctx context.Context) (ctrl.Manager, error) {
	webhookOpts := webhook.Options{
		Host:     f.webhookServerHost,
		Port:     f.webhookServerPort,
		CertDir:  f.webhookServerCertDir,
		CertName: f.webhookServerCertName,
		KeyName:  f.webhookServerKeyName,
	}
	webhookServer := webhook.NewServer(webhookOpts)

	return mgr.New(
		mgr.WithContext(ctx),
		mgr.WithAddToScheme(clientgoscheme.AddToScheme),
		mgr.WithAddToScheme(gardenercorev1beta1.AddToScheme),
		mgr.WithInstallScheme(configinstall.Install),
		mgr.WithMetricsAddress(f.metricsBindAddr),
		mgr.WithHealthProbeAddress(f.healthProbeBindAddr),
		mgr.WithLeaderElection(f.leaderElection),
		mgr.WithLeaderElectionID(f.leaderElectionID),
		mgr.WithLeaderElectionNamespace(f.leaderElectionNamespace),
		mgr.WithHealthzCheck("healthz", healthz.Ping),
		mgr.WithReadyzCheck("readyz", healthz.Ping),
		mgr.WithPprofAddress(f.pprofBindAddr),
		mgr.WithConnectionConfiguration(&componentbaseconfigv1alpha1.ClientConnectionConfiguration{
			QPS:   f.clientConnQPS,
			Burst: f.clientConnBurst,
		}),
		mgr.WithWebhookServer(webhookServer),
	)
}

// flagsKey is the key used to store the parsed command-line flags in a
// [context.Context].
type flagsKey struct{}

// getFlags extracts and returns the [flags] from the given [context.Context].
func getFlags(ctx context.Context) *flags {
	conf, ok := ctx.Value(flagsKey{}).(*flags)
	if !ok {
		return &flags{}
	}

	return conf
}

// NewWebhookCommand creates a new [cli.Command] for running the webhook server.
func NewWebhookCommand() *cli.Command {
	flags := flags{}

	cmd := &cli.Command{
		Name:    "webhook",
		Aliases: []string{"w"},
		Usage:   "start webhook server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "extension-name",
				Usage:       "name of the gardener extension",
				Value:       "gardener-extension-example-admission",
				Sources:     cli.EnvVars("EXTENSION_NAME"),
				Destination: &flags.extensionName,
			},
			&cli.StringFlag{
				Name:        "metrics-bind-address",
				Usage:       "the address the metrics endpoint binds to",
				Value:       ":8080",
				Sources:     cli.EnvVars("METRICS_BIND_ADDRESS"),
				Destination: &flags.metricsBindAddr,
			},
			&cli.StringFlag{
				Name:        "pprof-bind-address",
				Usage:       "the address at which pprof binds to",
				Sources:     cli.EnvVars("PPROF_BIND_ADDRESS"),
				Destination: &flags.pprofBindAddr,
			},
			&cli.StringFlag{
				Name:        "health-probe-bind-address",
				Usage:       "the address the probe endpoint binds to",
				Value:       ":8081",
				Sources:     cli.EnvVars("HEALTH_PROBE_BIND_ADDRESS"),
				Destination: &flags.healthProbeBindAddr,
			},
			&cli.BoolFlag{
				Name:        "leader-election",
				Usage:       "enable leader election for controller manager",
				Value:       false,
				Sources:     cli.EnvVars("LEADER_ELECTION"),
				Destination: &flags.leaderElection,
			},
			&cli.StringFlag{
				Name:        "leader-election-id",
				Usage:       "the leader election id to use, if leader election is enabled",
				Value:       "gardener-extension-example-admission",
				Sources:     cli.EnvVars("LEADER_ELECTION_ID"),
				Destination: &flags.leaderElectionID,
			},
			&cli.StringFlag{
				Name:        "leader-election-namespace",
				Usage:       "namespace to use for the leader election lease",
				Value:       "gardener-extension-example",
				Sources:     cli.EnvVars("LEADER_ELECTION_NAMESPACE"),
				Destination: &flags.leaderElectionNamespace,
			},
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "path to a kubeconfig when running out-of-cluster",
				Sources:     cli.EnvVars("KUBECONFIG"),
				Destination: &flags.kubeconfig,
				Action: func(ctx context.Context, c *cli.Command, val string) error {
					return os.Setenv(clientcmd.RecommendedConfigPathEnvVar, val)
				},
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Zap Level to configure the verbosity of logging",
				Value: glogger.InfoLevel,
				Validator: func(val string) error {
					if !slices.Contains(glogger.AllLogLevels, val) {
						return errors.New("invalid log level specified")
					}

					return nil
				},
				Destination: &flags.zapLogLevel,
			},
			&cli.StringFlag{
				Name:  "log-format",
				Usage: "Zap log encoding format, json or text",
				Value: glogger.FormatText,
				Validator: func(val string) error {
					if !slices.Contains(glogger.AllLogFormats, val) {
						return errors.New("invalid log level format specified")
					}

					return nil
				},
				Destination: &flags.zapLogFormat,
			},
			&cli.Float32Flag{
				Name:        "client-conn-qps",
				Usage:       "allowed client queries per second for the connection",
				Value:       100.0,
				Sources:     cli.EnvVars("CLIENT_CONNECTION_QPS"),
				Destination: &flags.clientConnQPS,
			},
			&cli.Int32Flag{
				Name:        "client-conn-burst",
				Usage:       "client connection burst size",
				Value:       130,
				Sources:     cli.EnvVars("CLIENT_CONNECTION_BURST"),
				Destination: &flags.clientConnBurst,
			},
			&cli.StringFlag{
				Name:        "gardener-version",
				Usage:       "version of gardener provided by gardenlet or gardener-operator",
				Sources:     cli.EnvVars("GARDENER_VERSION"),
				Destination: &flags.gardenerVersion,
			},
			&cli.BoolFlag{
				Name:        "self-hosted-shoot-cluster",
				Usage:       "set to true, if the extension runs in a self-hosted shoot cluster",
				Sources:     cli.EnvVars("SELF_HOSTED_SHOOT_CLUSTER"),
				Destination: &flags.selfHostedShootCluster,
			},
			&cli.StringFlag{
				Name:        "webhook-server-host",
				Usage:       "address on which the webhook server listens on",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_HOST"),
				Destination: &flags.webhookServerHost,
			},
			&cli.IntFlag{
				Name:        "webhook-server-port",
				Value:       9443,
				Usage:       "port on which the webhook server listens on",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_PORT"),
				Destination: &flags.webhookServerPort,
			},
			&cli.StringFlag{
				Name:        "webhook-server-cert-dir",
				Usage:       "path to directory, which contains the server key and cert",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_CERT_DIR"),
				Destination: &flags.webhookServerCertDir,
			},
			&cli.StringFlag{
				Name:        "webhook-server-cert-name",
				Value:       "tls.crt",
				Usage:       "the server certificate file name",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_CERT_NAME"),
				Destination: &flags.webhookServerCertName,
			},
			&cli.StringFlag{
				Name:        "webhook-server-key-name",
				Value:       "tls.key",
				Usage:       "the server certificate key file name",
				Sources:     cli.EnvVars("WEBHOOK_SERVER_KEY_NAME"),
				Destination: &flags.webhookServerKeyName,
			},
			&cli.StringFlag{
				Name:        "webhook-config-namespace",
				Usage:       "namespace where the webhook CA bundle, services, etc. are created",
				Sources:     cli.EnvVars("WEBHOOK_CONFIG_NAMESPACE"),
				Destination: &flags.webhookConfigNamespace,
			},
			&cli.StringFlag{
				Name:    "webhook-config-mode",
				Value:   string(extensionswebhook.ModeService),
				Usage:   "one of `service', `url' or `url-service'",
				Sources: cli.EnvVars("WEBHOOK_CONFIG_MODE"),
				Validator: func(val string) error {
					supportedModes := []string{
						string(extensionswebhook.ModeService),
						string(extensionswebhook.ModeURL),
						string(extensionswebhook.ModeURLWithServiceName),
					}
					if !slices.Contains(supportedModes, val) {
						return errors.New("invalid webhook config mode specified")
					}

					return nil
				},
				Destination: &flags.webhookConfigMode,
			},
			&cli.StringFlag{
				Name:    "webhook-config-url",
				Usage:   "URL at which to find the webhook server, used with `url' mode only",
				Sources: cli.EnvVars("WEBHOOK_CONFIG_URL"),
				Validator: func(val string) error {
					_, err := url.Parse(val)

					return err
				},
				Destination: &flags.webhookConfigURL,
			},
			&cli.IntFlag{
				Name:    "webhook-config-service-port",
				Usage:   "service port for the webhook when running in `service' mode",
				Sources: cli.EnvVars("WEBHOOK_CONFIG_SERVICE_PORT"),
				Validator: func(val int) error {
					if val <= 0 {
						return errors.New("port cannot be negative")
					}

					return nil
				},
				Destination: &flags.webhookConfigServicePort,
			},
			&cli.StringFlag{
				Name:        "webhook-config-owner-namespace",
				Usage:       "namespace which is used as the owner reference for webhook registration",
				Sources:     cli.EnvVars("WEBHOOK_CONFIG_OWNER_NAMESPACE"),
				Destination: &flags.webhookConfigOwnerNamespace,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			ctrllog.SetLogger(glogger.MustNewZapLogger(flags.zapLogLevel, flags.zapLogFormat))
			newCtx := context.WithValue(ctx, flagsKey{}, &flags)

			return newCtx, nil
		},
		Action: runWebhookServer,
	}

	return cmd
}

// runWebhookServer starts the webhook server.
func runWebhookServer(ctx context.Context, cmd *cli.Command) error {
	return nil
}
