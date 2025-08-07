package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/krateoplatformops/azuredevops-provider/internal/controllers"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/krateoplatformops/azuredevops-provider/apis"
	"github.com/krateoplatformops/azuredevops-provider/internal/controller-utils/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"

	"github.com/stoewer/go-strcase"
)

const (
	providerName = "Azure Devops"
)

func main() {
	envVarPrefix := fmt.Sprintf("%s_PROVIDER", strcase.UpperSnakeCase(providerName))

	var (
		app = kingpin.New(filepath.Base(os.Args[0]), fmt.Sprintf("Krateo %s Provider.", providerName)).
			DefaultEnvars()
		debug = app.Flag("debug", "Run with debug logging.").Short('d').
			OverrideDefaultFromEnvar(fmt.Sprintf("%s_DEBUG", envVarPrefix)).
			Bool()
		syncPeriod = app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').
				Default("1h").
				Duration()
		pollInterval = app.Flag("poll", "Poll interval controls how often an individual resource should be checked for drift.").
				Default("2m").
				OverrideDefaultFromEnvar(fmt.Sprintf("%s_POLL_INTERVAL", envVarPrefix)).
				Duration()
		maxReconcileRate = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").
					Default("5").
					OverrideDefaultFromEnvar(fmt.Sprintf("%s_MAX_RECONCILE_RATE", envVarPrefix)).
					Int()
		leaderElection = app.Flag("leader-election", "Use leader election for the controller manager.").
				Short('l').
				Default("false").
				OverrideDefaultFromEnvar(fmt.Sprintf("%s_LEADER_ELECTION", envVarPrefix)).
				Bool()
		maxErrorRetryInterval = app.Flag("max-error-retry-interval", "The maximum interval between retries when an error occurs. This should be less than the half of the poll interval.").
					Default("1m").
					OverrideDefaultFromEnvar(fmt.Sprintf("%s_MAX_ERROR_RETRY_INTERVAL", envVarPrefix)).
					Duration()
		minErrorRetryInterval = app.Flag("min-error-retry-interval", "The minimum interval between retries when an error occurs. This should be less than max-error-retry-interval.").
					Default("1s").
					OverrideDefaultFromEnvar(fmt.Sprintf("%s_MIN_ERROR_RETRY_INTERVAL", envVarPrefix)).
					Duration()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	//zl := zap.New(zap.UseDevMode(*debug))
	//log := logging.NewLogrLogger(zl.WithName(fmt.Sprintf("%s-provider", strcase.KebabCase(providerName))))
	//if *debug {
	//	// The controller-runtime runs with a no-op logger by default. It is
	//	// *very* verbose even at info level, so we only provide it a real
	//	// logger when we're running in debug mode.
	//}

	var zapOptions []zap.Opts
	if *debug {
		// Debug mode: mostra DEBUG, INFO, WARN, ERROR
		zapOptions = []zap.Opts{
			zap.UseDevMode(true),
			zap.Level(zapcore.DebugLevel),
		}
	} else {
		// Production mode: mostra solo INFO, WARN, ERROR
		zapOptions = []zap.Opts{
			zap.UseDevMode(false),
			zap.Level(zapcore.InfoLevel),
		}
	}
	zl := zap.New(zapOptions...)
	log := logging.NewLogrLogger(zl.WithName(fmt.Sprintf("%s-provider", strcase.KebabCase(providerName))))
	ctrl.SetLogger(zl)

	log.Debug("Starting", "sync-period", syncPeriod.String())

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:   *leaderElection,
		LeaderElectionID: fmt.Sprintf("leader-election-%s-provider", strcase.KebabCase(providerName)),
		Cache: cache.Options{
			SyncPeriod: syncPeriod,
		},
		Metrics: metricsserver.Options{
			BindAddress: ":8080",
		},
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	o := controller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcileRate,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       ratelimiter.NewGlobalExponential(*minErrorRetryInterval, *maxErrorRetryInterval),
	}

	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add APIs to scheme")
	kingpin.FatalIfError(controllers.Setup(mgr, o), "Cannot setup controllers")
	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
