package operator

import (
	"context"
	"github.com/go-logr/zapr"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.dfds.cloud/ssu-k8s/core/git"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/operator/controller"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme = runtime.NewScheme()
)

func InitOperator(ctx context.Context) {
	ctrl.SetLogger(zapr.NewLogger(logging.Logger))
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(rbacv1.AddToScheme(scheme))

	conf, err := config.LoadConfig()
	if err != nil {
		logging.Logger.Fatal("Failed to load config", zap.Error(err))
	}

	repo, err := git.LoadRepo(git.Config{
		RemoteRepoUri:     conf.Git.RemoteRepoURI,
		TemporaryRepoPath: conf.Git.TemporaryRepoPath,
		GitUsername:       "ssu-k8s",
		GitEmail:          "ssu-k8s@dfds.cloud",
		Branch:            conf.Git.Branch,
	})
	if err != nil {
		log.Fatal(err)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		//Metrics:          metricsServerOptions,
		Metrics: server.Options{
			BindAddress: "0", // disables metrics for now
		},
		LeaderElection:   false,
		LeaderElectionID: "beb9eb71.my.domain",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		logging.Logger.Fatal("unable to start manager", zap.Error(err))
	}

	if err = (&controller.NamespaceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Repo:   repo,
	}).SetupWithManager(mgr); err != nil {
		logging.Logger.Error("unable to create controller", zap.Error(err), zap.String("controller", "namespace"))
		os.Exit(1)
	}

	if err = (&controller.SecretReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Repo:   repo,
	}).SetupWithManager(mgr); err != nil {
		logging.Logger.Error("unable to create controller", zap.Error(err), zap.String("controller", "secret"))
		os.Exit(1)
	}

	logging.Logger.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		logging.Logger.Error("problem running manager", zap.Error(err))
		os.Exit(1)
	}
}
