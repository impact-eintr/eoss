package temp

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/errmsg"
)

func Put(ctx *gin.Context) {
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

	info, err := f.Stat()
	if err != nil {
		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	actual := info.Size()
	os.Remove(infoFile)
	if actual != tempinfo.Size {
		os.Remove(datFile)
		log.Println("actual size", actual, "exceeds", tempinfo.Size)

		errmsg.ErrRawLog(ctx, http.StatusInternalServerError, "tempfile mismatch size")
		return
	}
	commitTempObject(datFile, tempinfo)
}
