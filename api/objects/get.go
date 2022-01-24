package objects

import (
	"fmt"
	"io"
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

func Get(c *gin.Context) {
	name := c.Param("name")
	versionId := c.Query("version")
	version := 0
	var e error
	if len(versionId) != 0 {
		version, e = strconv.Atoi(versionId)
		if e != nil {
			errmsg.ErrLog(c, http.StatusBadRequest, e.Error())
			return
		}
	}
	meta, e := es.GetMetadata(name, version)
	if e != nil {
		errmsg.ErrLog(c, http.StatusInternalServerError, e.Error())
		return
	}

	if meta.Hash == "" {
		errmsg.ErrLog(c, http.StatusNotFound, "File not found")
		return
	}

	hash := url.PathEscape(meta.Hash) // 这个是转为URL后的hash_value
	stream, e := GetStream(hash, meta.Size)
	if e != nil {
		errmsg.ErrLog(c, http.StatusNotFound, e.Error())
		return
	}

	offset := utils.GetOffsetFromHeader(c.Request.Header)
	if offset != 0 {
		stream.Seek(offset, io.SeekCurrent)
		c.Writer.Header().Set("content-range", fmt.Sprintf("bytes %d-%d/%d", offset, meta.Size-1, meta.Size))
		c.Writer.WriteHeader(http.StatusPartialContent)
	}

	if _, err := io.Copy(c.Writer, stream); err != nil {
		errmsg.ErrLog(c, http.StatusOK, err.Error())
		return
	}
	stream.Close()
}

func GetStream(hash string, size int64) (*rs.RSGetStream, error) {
	locateInfo := locate.Locate(hash)
	if len(locateInfo) < rs.DATA_SHARDS {
		return nil, fmt.Errorf("object %s locate fail, result %v", hash, locateInfo)
	}
	dataServers := make([]string, 0)
	if len(locateInfo) != rs.ALL_SHARDS {
		dataServers = heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS-len(locateInfo), locateInfo)
	}
	return rs.NewRSGetStream(locateInfo, dataServers, hash, size)
}
