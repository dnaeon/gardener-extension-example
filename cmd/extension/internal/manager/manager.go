// TODO: requeue interval flag

package manager

import (
	"context"
	"errors"
	"slices"
	"time"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	glogger "github.com/gardener/gardener/pkg/logger"
	"github.com/urfave/cli/v3"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// flags stores the manager flags as provided from the command-line
type flags struct {
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
	extensionClass            string
	zapLogLevel               string
	zapLogFormat              string
	resyncInterval            time.Duration
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
		Name:  "manager",
		Usage: "start controller manager",
		Flags: []cli.Flag{
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
				Name:     "extension-class",
				Usage:    "extension class this extension is responsible for [garden, seed or shoot]",
				Required: true,
				Sources:  cli.EnvVars("EXTENSION_CLASS"),
				Validator: func(val string) error {
					classes := []string{
						string(extensionsv1alpha1.ExtensionClassGarden),
						string(extensionsv1alpha1.ExtensionClassSeed),
						string(extensionsv1alpha1.ExtensionClassShoot),
					}
					if !slices.Contains(classes, val) {
						return errors.New("invalid extension class specified")
					}

					return nil
				},
				Destination: &flags.extensionClass,
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
				Usage: "Zap log encoding format, json or console",
				Value: glogger.FormatJSON,
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
	// TODO: fill me in

	return nil
}
