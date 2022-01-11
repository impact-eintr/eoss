package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/data/heartbeat"
	"github.com/impact-eintr/eoss/data/locate"
	"github.com/impact-eintr/eoss/data/objects"
)

func main() {
	go heartbeat.StartHeartbeat()
	go locate.StartLocate()

	eng := gin.Default()

	objGroup := eng.Group("/objects")
	{
		objGroup.PUT("/*filehash", objects.Put)
		objGroup.GET("/*filehash", objects.Get)
		//objGroup.DELETE("/*filehash", objects.Delete)
	}

	eng.Run(":" + os.Getenv("LISTEN_PORT"))
}
