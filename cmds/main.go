package main

import (
	"go.dfds.cloud/bootstrap"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/api"
	"go.dfds.cloud/ssu-k8s/feats/jobs"
	"go.dfds.cloud/ssu-k8s/feats/operator"
	"go.uber.org/zap"
	"log"
)

func main() {
	// setup base
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	builder := bootstrap.Builder()
	builder.EnableLogging(conf.Log.Debug, conf.Log.Level)
	builder.EnableHttpRouter(conf.Http.Enabled)
	if conf.Metrics.Enabled {
		builder.EnableMetrics()
	}
	builder.EnableOrchestrator("orchestrator")
	manager := builder.Build()
	logging.Logger = manager.Logger
	manager.Orchestrator.Init(logging.Logger)

	logging.Logger.Info("ssu-k8s launched")

	// setup feats
	api.Configure(manager.HttpRouter)

	jobs.Init(manager.Orchestrator)
	//msgWg := messaging.Init(manager)

	// run
	go operator.InitOperator()
	<-manager.Context.Done()
	if err := manager.HttpServer.Shutdown(manager.Context); err != nil {
		logging.Logger.Info("HTTP Server was unable to shut down gracefully", zap.Error(err))
	}

	//msgWg.Wait()

	logging.Logger.Info("server shutting down")

}
