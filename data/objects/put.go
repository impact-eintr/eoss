package objects

import (
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
)

func Put(ctx *gin.Context) {
	path := "/tmp/eoss/objects/" + url.PathEscape(ctx.Param("filehash")[1:])
	f, err := os.Create(path)
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, "file create faild")
		return
	}
	defer f.Close()
	io.Copy(f, ctx.Request.Body)
}
