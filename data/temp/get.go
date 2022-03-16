package temp

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
)

func Get(c *gin.Context) {
	uuid := c.Param("uuid")
	f, e := os.Open("/tmp/eoss/temp/" + uuid + ".dat")
	if e != nil {
		errmsg.ErrLog(c, http.StatusNotFound, e.Error())
		return
	}
	defer f.Close()
	io.Copy(c.Writer, f)
}
