package main

import (
	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/objects"
)

func main() {
	eng := gin.Default()

	objGroup := eng.Group("/objects")
	{
		objGroup.PUT("/", objects.Put)
		objGroup.GET("/", objects.Get)
		objGroup.DELETE("/", objects.Delete)
	}

	locGroup := eng.Group("/locate")
	{

	}
}
