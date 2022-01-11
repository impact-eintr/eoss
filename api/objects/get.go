package objects

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/heartbeat"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
	"github.com/impact-eintr/eoss/rs"
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
		errmsg.ErrLog(c, http.StatusNotFound, e.Error())
		return
	}

	hash := url.PathEscape(meta.Hash) // 这个是转为URL后的hash_value
	stream, e := GetStream(hash, meta.Size)
	if e != nil {
		errmsg.ErrLog(c, http.StatusNotFound, e.Error())
		return
	}

	if _, err := io.Copy(c.Writer, ioutil.NopCloser(stream)); err != nil {
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
