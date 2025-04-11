package util

import (
	"fmt"
	"go.dfds.cloud/ssu-k8s/core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitializeLogger() {
	conf, _ := config.LoadConfig()

	var logConf zap.Config
	if conf.LogDebug {
		logConf = zap.NewDevelopmentConfig()
	} else {
		logConf = zap.NewProductionConfig()
	}

	level, err := zapcore.ParseLevel(conf.LogLevel)
	if err != nil {
		fmt.Println(err)
		level = zapcore.InfoLevel
	}

	logConf.Level = zap.NewAtomicLevelAt(level)

	Logger, _ = logConf.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	Logger.Info(fmt.Sprintf("Logging enabled, log level set to %s", Logger.Level().String()))
}
