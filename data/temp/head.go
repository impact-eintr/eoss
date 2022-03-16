package temp

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
)

func Head(c *gin.Context) {
	uuid := c.Param("uuid")
	f, e := os.Open("/tmp/eoss/temp/" + uuid + ".dat")
	if e != nil {
		errmsg.ErrRawLog(c, http.StatusNotFound, e.Error())
		return
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		errmsg.ErrRawLog(c, http.StatusInternalServerError, e.Error())
		return
	}
	c.Writer.Header().Set("content-length", fmt.Sprintf("%d", info.Size()))
}
