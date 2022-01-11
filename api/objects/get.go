package objects

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
	"github.com/impact-eintr/eoss/objectstream"
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

	// meta.Hash 是不能直接用作URL的
	if meta.Hash == "" {
		c.Status(http.StatusNotFound)
		return
	}

	object := url.PathEscape(meta.Hash) // 这个是转为URL后的hash_value
	stream, e := getStream(object)
	if e != nil {
		errmsg.ErrLog(c, http.StatusNotFound, e.Error())
		return
	}

	//data, _ := ioutil.ReadAll(stream)
	//c.Data(http.StatusOK, "application/octet-stream", data)
	io.Copy(c.Writer, ioutil.NopCloser(stream))
}

func getStream(object string) (io.Reader, error) {
	server := locate.Locate(object)
	if server == "" {
		return nil, fmt.Errorf("object %s locate fail", object)

	}
	return objectstream.NewGetStream(server, object)
}
