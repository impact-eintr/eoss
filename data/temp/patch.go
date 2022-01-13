package temp

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
)

func Patch(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	tempinfo, err := readFromFile(uuid)
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusNotFound, err.Error())
		return
	}

	infoFile := "/tmp/eoss/temp/" + uuid
	datFile := infoFile + ".dat"
	f, err := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer f.Close()

	_, err = io.Copy(f, ctx.Request.Body)
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	info, err := f.Stat()
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	actual := info.Size()
	if actual > tempinfo.Size {
		os.Remove(datFile)
		os.Remove(infoFile)
		log.Println("actual size", actual, "exceeds", tempinfo.Size)
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, "tempfile mismatch size")
		return
	}
}

func readFromFile(uuid string) (*tempInfo, error) {
	f, e := os.Open("/tmp/eoss/temp/" + uuid) // 读取临时文件的元数据
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	var info tempInfo
	json.Unmarshal(b, &info)
	return &info, nil
}
