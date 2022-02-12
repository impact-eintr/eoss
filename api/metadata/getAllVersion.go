package metadata

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/es"
)

// 获取单个文件的所有版本
func GetAllVersions(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		return
	}
	res, err := es.SearchAllVersions(name, 0, 10)
	if err != nil {
		log.Println(err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// 获得所有文件的最新版本
func GetLastestVersions(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, es.LatestMetadatas())
}
