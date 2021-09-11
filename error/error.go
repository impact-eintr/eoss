package error

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrMsg struct {
	Code int
	Msg  string
}

// code 为业务逻辑错误码 http响应统一返回200
func ErrLog(ctx *gin.Context, code int, msg string) {
	errMsg := ErrMsg{
		Code: code,
		Msg:  msg,
	}

	log.Println(msg)

	ctx.JSON(http.StatusOK, errMsg)
}
