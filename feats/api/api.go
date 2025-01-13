package api

import (
	"github.com/gin-gonic/gin"
	"go.dfds.cloud/ssu-k8s/feats/api/controller"
)

func Configure(router *gin.Engine) {
	controller.AddControllers(router)
}
