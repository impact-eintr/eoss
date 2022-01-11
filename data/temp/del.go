package temp

import (
	"os"

	"github.com/gin-gonic/gin"
)

func Del(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	infoFile := "/tmp/eoss/temp/" + uuid
	datFile := infoFile + ".dat"
	os.Remove(infoFile)
	os.Remove(datFile)
}
