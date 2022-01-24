package objects

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/es"
)

func Post(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, es.LatestMetadatas())
}
