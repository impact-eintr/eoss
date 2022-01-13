package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/heartbeat"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/api/objects"
	"github.com/impact-eintr/eoss/mq/esqv1"
)

func main() {
	go heartbeat.ListenHeartbeat() // 监听心跳
	go esqv1.ListenFileResponse()  // 监听文件信息

	eng := gin.Default()

	objGroup := eng.Group("/objects")
	{
		objGroup.POST("/:name", objects.Post)
		objGroup.PUT("/:name", objects.Put)
		objGroup.GET("/:name", objects.Get)
		objGroup.DELETE("/:name", objects.Delete)
	}

	eng.Group("/locate", locate.Get)

	eng.Run(":" + os.Getenv("LISTEN_PORT"))
}
