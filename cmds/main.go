package main

import (
	"go.dfds.cloud/bootstrap"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/api"
	"go.dfds.cloud/ssu-k8s/feats/jobs"
	"go.dfds.cloud/ssu-k8s/feats/messaging"
	"go.dfds.cloud/ssu-k8s/feats/operator"
	"go.uber.org/zap"
	"sync"
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

	conf, err := config.LoadConfig()
	if err != nil {
		logging.Logger.Fatal("failed to load config", zap.Error(err))
	}

	// setup graceful shutdown
	// trap Ctrl+C and call cancel on the context

	// setup feats
	api.Configure(manager.HttpRouter)

	jobs.Init(manager.Orchestrator)

	var msgWg *sync.WaitGroup
	if conf.Enable.Messaging {
		_, wg := messaging.Init(manager)
		msgWg = wg
	}

	// run
	if conf.Enable.Operator {
		go operator.InitOperator(manager.Context)
	}

	<-manager.Context.Done()
	if err := manager.HttpServer.Shutdown(manager.Context); err != nil {
		logging.Logger.Info("HTTP Server was unable to shut down gracefully", zap.Error(err))
	}

	if conf.Enable.Messaging && msgWg != nil {
		msgWg.Wait()
	}

	logging.Logger.Info("server shutting down")
}
