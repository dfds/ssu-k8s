package main

import (
	"go.dfds.cloud/bootstrap"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/api"
	"go.dfds.cloud/ssu-k8s/feats/jobs"
	"go.dfds.cloud/ssu-k8s/feats/messaging"
	"go.dfds.cloud/ssu-k8s/feats/operator"
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

	jobs.Init(manager.Orchestrator)
	msgWg := messaging.Init(manager)

	// run
	go operator.InitOperator()
	<-manager.Context.Done()
	if err := manager.HttpServer.Shutdown(manager.Context); err != nil {
		logging.Logger.Info("HTTP Server was unable to shut down gracefully", zap.Error(err))
	}

	msgWg.Wait()

	logging.Logger.Info("server shutting down")

}
