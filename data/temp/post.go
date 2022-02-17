package temp

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/impact-eintr/eoss/data/locate"
	"github.com/impact-eintr/eoss/errmsg"
)

type tempInfo struct {
	Uuid string
	Name string
	Size int64
}

func Post(ctx *gin.Context) {
	uuid := uuid.New().String()
	name := url.PathEscape(ctx.Param("filehash")[1:])

	if locate.Locate(name) != -1 {
		// TODO 告诉api消息有延迟 不要发了
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	size, err := strconv.ParseInt(ctx.Request.Header.Get("size"), 0, 64)
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	t := tempInfo{uuid, name, size}
	err = t.writeToFile()
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	os.Create("/tmp/eoss/temp/" + t.Uuid + ".dat") // 创造空的临时文件
	ctx.Writer.Write([]byte(uuid))
}

func (t *tempInfo) writeToFile() error {
	f, err := os.Create("/tmp/eoss/temp/" + t.Uuid) // 写入临时文件元数据
	if err != nil {
		return err
	}
	defer f.Close()
	b, _ := json.Marshal(t)
	f.Write(b)
	return nil
}
