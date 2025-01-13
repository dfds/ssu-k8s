package main

import (
	"context"
	"go.dfds.cloud/bootstrap"
	"go.dfds.cloud/orchestrator"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/api"
	"go.dfds.cloud/ssu-k8s/feats/messaging"
	"go.uber.org/zap"
)

func main() {
	// setup base
	builder := bootstrap.Builder()
	builder.EnableLogging(false, "debug")
	builder.EnableHttpRouter(false)
	builder.EnableMetrics()
	builder.EnableOrchestrator("orchestrator")
	manager := builder.Build()
	logging.Logger = manager.Logger
	manager.Orchestrator.Init(logging.Logger)

	logging.Logger.Info("ssu-k8s launched")

	// setup feats
	api.Configure(manager.HttpRouter)

	configPrefix := "SSU_K8S_JOB"
	manager.Orchestrator.AddJob(configPrefix, orchestrator.NewJob("dummy", func(ctx context.Context) error {
		logging.Logger.Info("dummy")
		return nil
	}), &orchestrator.Schedule{})

	manager.Orchestrator.Run()

	msgWg := messaging.Init(manager)

	// run
	<-manager.Context.Done()
	if err := manager.HttpServer.Shutdown(manager.Context); err != nil {
		logging.Logger.Info("HTTP Server was unable to shut down gracefully", zap.Error(err))
	}

	msgWg.Wait()

	logging.Logger.Info("server shutting down")

}
