package objects

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/heartbeat"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
	"github.com/impact-eintr/eoss/rs"
	"github.com/impact-eintr/eoss/utils"
)

func Post(c *gin.Context) {
	name := c.Param("name")
	size, e := strconv.ParseInt(c.Request.Header.Get("size"), 0, 64)
	if e != nil {
		errmsg.ErrLog(c, http.StatusForbidden, e.Error())
		return
	}
	hash := utils.GetHashFromHeader(c.Request.Header)
	if hash == "" {
		errmsg.ErrLog(c, http.StatusBadRequest, e.Error())
		return
	}
	location := utils.GetLocationFromHeader(c.Request.Header)
	if location == "" {
		errmsg.ErrLog(c, http.StatusBadRequest, e.Error())
		return
	}
	if locate.Exist(url.PathEscape(hash)) {
		e = es.AddVersion(name, hash, location, size)
		if e != nil {
			errmsg.ErrLog(c, http.StatusInternalServerError, e.Error())
		} else {
			c.Status(http.StatusOK)
		}
		return
	}
	// 找到需要数量的服务器
	ds := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)
	if len(ds) != rs.ALL_SHARDS {
		errmsg.ErrLog(c, http.StatusServiceUnavailable, e.Error())
		return
	}
	stream, e := rs.NewRSResumablePutStream(ds, name, url.PathEscape(hash), location, size)
	if e != nil {
		errmsg.ErrLog(c, http.StatusInternalServerError, e.Error())
		return
	}
	c.Writer.Header().Set("location", "/temp/"+url.PathEscape(stream.ToToken()))
	c.Status(http.StatusCreated)
}
