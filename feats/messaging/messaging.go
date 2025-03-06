package messaging

import (
	"go.dfds.cloud/bootstrap"
	"go.dfds.cloud/messaging"
	"go.dfds.cloud/ssu-k8s/core/logging"
	"go.dfds.cloud/ssu-k8s/feats/messaging/handlers"
	"go.uber.org/zap"
	"sync"
)

func Init(manager *bootstrap.Manager) *sync.WaitGroup {
	msgWg := &sync.WaitGroup{}
	msg := messaging.CreateMessaging()
	err := msg.Init(manager.Context, &messaging.Config{
		EnvVarPrefix: "SSU_K8S_KAFKA_AUTH",
		Wg:           msgWg,
		Logger:       logging.Logger,
	})
	if err != nil {
		logging.Logger.Fatal("Failed to init messaging", zap.Error(err))
	}

	configure(msg)

	return msgWg
}

func configure(msg *messaging.Messaging) {
	auditConsumer := msg.NewConsumer("build.selfservice.events.capabilities", "cloudengineering.ssu-k8s")
	auditConsumer.Register("aws_context_account_created", handlers.AwsContextAccountCreatedHandler)
	
	go auditConsumer.StartConsumer()
}
