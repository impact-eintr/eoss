package objects

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/heartbeat"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
	"github.com/impact-eintr/eoss/rs"
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

	// 先准备存数据
	c, err := storeObject(ctx.Request.Body, hash, size)
	if err != nil {
		errmsg.ErrLog(ctx, c, err.Error())
		return
	}
	if c != http.StatusOK {
		errmsg.ErrLog(ctx, c, "未知错误")
		return
	}

	// 再存元数据 注意如果新上传的文件相较上一个版本没有改变 就不会更新实际存在的文件 只更新元数据
	err = es.AddVersion(name, hash, size)
	if err != nil {
		errmsg.ErrLog(ctx, http.StatusInternalServerError, err.Error())
		return
	}

}

// r: 文件流
// object: 将文件hash后得到的hash_string
func storeObject(r io.Reader, hash string, size int64) (int, error) {
	if locate.Exist(url.PathEscape(hash)) {
		return http.StatusOK, nil
	}

	stream, err := putStream(url.PathEscape(hash), size)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	reader := io.TeeReader(r, stream)
	d := utils.CalculateHash(reader)
	if d != hash {
		stream.Commit(false)
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch, calculated=%s, request=%s", d, hash)
	}
	// 这里调用了 RSPutStream.Flush() 然后调用了 TempPutStream.Write()
	// 向 dataServers 发送了 PATCH 请求 将分片数据发送给了 dataServers
	// 如果 PATH 请求成功的话 会继续调用 TempPutStream.Commit()
	// 同时会根据传入的 true/false 判断继续向 dataServers 发送 PUT/DELETE 请求 来将临时文件 转正/删除
	stream.Commit(true)

	return http.StatusOK, nil
}

func putStream(hash string, size int64) (*rs.RSPutStream, error) {
	servers := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)
	if len(servers) != rs.ALL_SHARDS {
		return nil, fmt.Errorf("cannot find enough dataServer")
	}
	return rs.NewRSPutStream(servers, hash, size)
}
