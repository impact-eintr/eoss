package temp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/rs"
)

func Head(c *gin.Context) {
	token := c.Param("token")
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		errmsg.ErrLog(c, http.StatusForbidden, e.Error())
		return
	}
	current := stream.CurrentSize()
	if current == -1 {
		errmsg.ErrLog(c, http.StatusNotFound, "this file not found")
		return
	}
	c.Writer.Header().Set("content-length", fmt.Sprintf("%d", current))
}
