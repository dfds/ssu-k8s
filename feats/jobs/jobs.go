package jobs

import (
	"context"
	"go.dfds.cloud/orchestrator"
	"go.dfds.cloud/ssu-k8s/core/logging"
)

func Init(orc *orchestrator.Orchestrator) {
	configPrefix := "SSU_K8S_JOB"

	orc.AddJob(configPrefix, orchestrator.NewJob("dummy", func(ctx context.Context) error {
		logging.Logger.Info("dummy")
		return nil
	}), &orchestrator.Schedule{})

	orc.Run()
}
