package jobs

import (
	"context"
	"go.dfds.cloud/orchestrator"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/jobs/handlers"
)

func Init(orc *orchestrator.Orchestrator) {
	configPrefix := "SSU_K8S_JOB"

	orc.AddJob(configPrefix, orchestrator.NewJob("dummy", func(ctx context.Context) error {
		logging.Logger.Info("dummy")
		return nil
	}), &orchestrator.Schedule{})

	orc.AddJob(configPrefix, orchestrator.NewJob("cacheKubernetesResources", handlers.CacheKubernetesResources), &orchestrator.Schedule{})

	orc.AddJob(configPrefix, orchestrator.NewJob("migrateExistingNamespaces", handlers.MigrateExistingNamespaces), &orchestrator.Schedule{})

	orc.Run()
}
