package objects

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
)

func Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	version, err := es.SearchLatestVersion(name)
	if err != nil {
		errmsg.ErrLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	err = es.PutMetadata(name, version.Version+1, 0, "", "")
	if err != nil {
		errmsg.ErrLog(ctx, http.StatusInternalServerError, err.Error())
	}
}
