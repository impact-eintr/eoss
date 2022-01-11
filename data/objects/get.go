package objects

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
)

func Get(ctx *gin.Context) {
	path := "/tmp/eoss/objects/" + url.PathEscape(ctx.Param("filehash")[1:])
	log.Println(path)
	f, err := os.Open(path)
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusNotFound, "cannot find file")
		return
	}
	defer f.Close()
	io.Copy(ctx.Writer, f)
}
