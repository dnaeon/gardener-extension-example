package manager

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/urfave/cli/v3"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"gardener-extension-example/pkg/actuator"
	"gardener-extension-example/pkg/controller"
	"gardener-extension-example/pkg/heartbeat"
	"gardener-extension-example/pkg/mgr"
)

// flags stores the manager flags as provided from the command-line
type flags struct {
	extensionName             string
	metricsBindAddr           string
	healthProbeBindAddr       string
	heartbeatRenewInterval    time.Duration
	heartbeatNamespace        string
	leaderElection            bool
	leaderElectionID          string
	leaderElectionNamespace   string
	ignoreOperationAnnotation bool
	maxConcurrentReconciles   int
	kubeconfig                string
	zapLogLevel               string
	zapLogFormat              string
	resyncInterval            time.Duration
}

// getManager creates a new [ctrl.Manager] based on the parsed [flags].
func (f *flags) getManager(ctx context.Context) (ctrl.Manager, error) {
	mgr, err := mgr.New(
		mgr.WithContext(ctx),
		mgr.WithAddToScheme(clientgoscheme.AddToScheme),
		mgr.WithAddToScheme(extensionscontroller.AddToScheme),
		mgr.WithMetricsAddress(f.metricsBindAddr),
		mgr.WithHealthProbeAddress(f.healthProbeBindAddr),
		mgr.WithLeaderElection(f.leaderElection),
		mgr.WithLeaderElectionID(f.leaderElectionID),
		mgr.WithLeaderElectionNamespace(f.leaderElectionNamespace),
		mgr.WithMaxConcurrentReconciles(f.maxConcurrentReconciles),
		mgr.WithHealthzCheck("healthz", healthz.Ping),
		mgr.WithReadyzCheck("readyz", healthz.Ping),
	)

	if err != nil {
		return nil, err
	}

	hb, err := heartbeat.New(
		heartbeat.WithExtensionName(f.extensionName),
		heartbeat.WithLeaseNamespace(f.heartbeatNamespace),
		heartbeat.WithRenewInterval(f.heartbeatRenewInterval),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create heartbeat controller: %w", err)
	}

	if err := hb.SetupWithManager(ctx, mgr); err != nil {
		return nil, fmt.Errorf("failed to setup heartbeat controller: %w", err)
	}

	return mgr, nil
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

// New creates a new [cli.Command] for running the controller manager.
func New() *cli.Command {
	var flags flags

	cmd := &cli.Command{
		Name:    "manager",
		Aliases: []string{"m"},
		Usage:   "start controller manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "extension-name",
				Usage:       "name of the gardener extension",
				Value:       "gardener-extension-example",
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
				Name:        "health-probe-bind-address",
				Usage:       "the address the probe endpoint binds to",
				Value:       ":8081",
				Sources:     cli.EnvVars("HEALTH_PROBE_BIND_ADDRESS"),
				Destination: &flags.healthProbeBindAddr,
			},
			&cli.DurationFlag{
				Name:        "heartbeat-renew-interval",
				Usage:       "renew heartbeat lease on specified interval",
				Value:       time.Duration(30 * time.Second),
				Sources:     cli.EnvVars("HEARTBEAT_RENEW_INTERVAL"),
				Destination: &flags.heartbeatRenewInterval,
			},
			&cli.StringFlag{
				Name:        "heartbeat-namespace",
				Usage:       "namespace to use for the heartbeat lease",
				Value:       "gardener-extension-example",
				Sources:     cli.EnvVars("HEARTBEAT_NAMESPACE"),
				Destination: &flags.heartbeatNamespace,
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
				Value:       "gardener-extension-example",
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
			&cli.BoolFlag{
				Name:        "ignore-operation-annotation",
				Usage:       "specifies whether to ignore operation annotation",
				Value:       false,
				Sources:     cli.EnvVars("IGNORE_OPERATION_ANNOTATION"),
				Destination: &flags.ignoreOperationAnnotation,
			},
			&cli.IntFlag{
				Name:        "max-concurrent-reconciles",
				Usage:       "max number of concurrent reconciliations",
				Value:       5,
				Sources:     cli.EnvVars("MAX_CONCURRENT_RECONCILES"),
				Destination: &flags.maxConcurrentReconciles,
			},
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "path to a kubeconfig when running out-of-cluster",
				Sources:     cli.EnvVars("KUBECONFIG"),
				Destination: &flags.kubeconfig,
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
			&cli.DurationFlag{
				Name:        "resync-interval",
				Usage:       "requeue interval of the controllers",
				Value:       time.Duration(30 * time.Second),
				Sources:     cli.EnvVars("RESYNC_INTERVAL"),
				Destination: &flags.resyncInterval,
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			ctrllog.SetLogger(glogger.MustNewZapLogger(flags.zapLogLevel, flags.zapLogFormat))
			newCtx := context.WithValue(ctx, flagsKey{}, &flags)

			return newCtx, nil
		},
		Action: runManager,
	}

	return cmd
}

// runManager starts the controller manager
func runManager(ctx context.Context, cmd *cli.Command) error {
	logger := ctrllog.Log.WithName("manager-setup")
	logger.Info("creating manager")

	flags := getFlags(ctx)
	mgr, err := flags.getManager(ctx)
	if err != nil {
		return err
	}

	logger.Info("creating actuators")
	act, err := actuator.New(
		actuator.WithReader(mgr.GetAPIReader()),
		actuator.WithClient(mgr.GetClient()),
	)
	if err != nil {
		return fmt.Errorf("failed to create actuator: %w", err)
	}

	logger.Info("creating controllers")
	c, err := controller.New(
		controller.WithActuator(act),
		controller.WithName(act.Name()),
		controller.WithExtensionType(act.ExtensionType()),
		controller.WithFinalizerSuffix(act.FinalizerSuffix()),
		controller.WithExtensionClasses(act.ExtensionClasses()),
		controller.WithIgnoreOperationAnnotation(flags.ignoreOperationAnnotation),
		controller.WithResyncInterval(flags.resyncInterval),
	)
	if err != nil {
		return fmt.Errorf("failed to create a controller: %w", err)
	}

	if err := c.SetupWithManager(ctx, mgr); err != nil {
		return fmt.Errorf("failed to setup controller with manager: %w", err)
	}

	logger.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}

	return nil
}
