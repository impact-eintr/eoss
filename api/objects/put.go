package objects

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/heartbeat"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
	"github.com/impact-eintr/eoss/objectstream"
	"github.com/impact-eintr/eoss/utils"
)

func Put(ctx *gin.Context) {
	// 获取文件 hash
	hash := utils.GetHashFromHeader(ctx.Request.Header)
	if hash == "" {
		errmsg.ErrLog(ctx, http.StatusBadRequest, "missing object hash int digest header")
		return
	}
	// 获取文件 size
	size := utils.GetSizeFromHeader(ctx.Request.Header)
	// 获取文件 name
	name := ctx.Param("name")

	// 先存元数据
	err := es.AddVersion(name, hash, size)
	if err != nil {
		errmsg.ErrLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 存数据
	c, err := storeObject(ctx.Request.Body, url.PathEscape(hash))
	if err != nil {
		errmsg.ErrLog(ctx, c, err.Error())
		return
	}
	if c != http.StatusOK {
		errmsg.ErrLog(ctx, c, "未知错误")
	}
}

// r: 文件流
// object: 将文件hash后得到的hash_string
func storeObject(r io.Reader, object string) (int, error) {
	stream, err := putStream(object)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	io.Copy(stream, r)
	err = stream.Close()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func putStream(object string) (*objectstream.PutStream, error) {
	server := heartbeat.ChooseRandomDataServer()
	if server == "" {
		return nil, fmt.Errorf("cannot find any dataServer")
	}
	return objectstream.NewPutStream(server, object), nil
}
