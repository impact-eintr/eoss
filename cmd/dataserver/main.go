package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/data/heartbeat"
	"github.com/impact-eintr/eoss/data/locate"
	"github.com/impact-eintr/eoss/data/objects"
	"github.com/impact-eintr/eoss/data/temp"
)

func init() {
	err := os.MkdirAll("/tmp/eoss/objects", os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	err = os.MkdirAll("/tmp/eoss/temp", os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	locate.CollectObjects()
	go heartbeat.StartHeartbeat()
	go locate.StartLocate()

	eng := gin.Default()

	objGroup := eng.Group("/objects")
	{
		objGroup.GET("/*filehash", objects.Get)
	}

	tmpGroup := eng.Group("/temp")
	{
		tmpGroup.POST("/*filehash", temp.Post)
		tmpGroup.PUT("/:uuid", temp.Put)
		tmpGroup.PATCH("/:uuid", temp.Patch)
		tmpGroup.DELETE("/:uuid", temp.Del)
	}

	eng.Run(":" + os.Getenv("LISTEN_PORT"))
}
