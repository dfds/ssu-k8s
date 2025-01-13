package main

import (
	"go.dfds.cloud/bootstrap"
	"go.uber.org/zap"
)

func main() {
	builder := bootstrap.Builder()
	builder.EnableLogging(false, "trace")
	builder.EnableHttpRouter(false)
	builder.EnableMetrics()
	builder.EnableOrchestrator("orchestrator")
	manager := builder.Build()

	manager.Logger.Info("ssu-k8s launched")
	<-manager.Context.Done()

	if err := manager.HttpServer.Shutdown(manager.Context); err != nil {
		manager.Logger.Info("HTTP Server was unable to shut down gracefully", zap.Error(err))
	}

	manager.Logger.Info("server shutting down")

}
