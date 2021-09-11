package objects

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/error"
)

func Put(ctx *gin.Context) {

	hash := utils.GetHansFromHeader(ctx.Request.Header)
	if hash == "" {
		error.ErrLog(ctx, http.StatusBadRequest, "missing object hash int digest header")
		return
	}

	size := utils.GetSizeFromHeader(ctx.Request.Header)

}
