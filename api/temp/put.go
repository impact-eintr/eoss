package temp

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/api/locate"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/es"
	"github.com/impact-eintr/eoss/rs"
	"github.com/impact-eintr/eoss/utils"
)

func Put(c *gin.Context) {
	token := c.Param("token")
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		errmsg.ErrLog(c, http.StatusForbidden, e.Error())
		return
	}
	current := stream.CurrentSize()
	if current == -1 {
		c.Status(http.StatusNotFound)
		return
	}
	offset := utils.GetOffsetFromHeader(c.Request.Header)
	if current != offset {
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	bytes := make([]byte, rs.BLOCK_SIZE)
	for {
		n, e := io.ReadFull(c.Request.Body, bytes)
		if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
			errmsg.ErrLog(c, http.StatusInternalServerError, e.Error())
			return
		}
		current += int64(n)        // 累计传输字节数
		if current > stream.Size { // 累计值非法
			stream.Commit(false)
			log.Println("resumable put exceed size")
			errmsg.ErrLog(c, http.StatusForbidden, e.Error())
			return
		}
		if n != rs.BLOCK_SIZE && current != stream.Size {
			return
		}
		stream.Write(bytes[:n])
		if current == stream.Size { // 累计值达到文件大小 结束循环 开始传输
			stream.Flush()
			getStream, e := rs.NewRSResumableGetStream(stream.Servers, stream.Uuids, stream.Size)
			hash := url.PathEscape(utils.CalculateHash(getStream))
			if hash != stream.Hash {
				stream.Commit(false)
				log.Println("resumable put done but hash mismatch")
				errmsg.ErrLog(c, http.StatusForbidden, e.Error())
				return
			}
			if locate.Exist(url.PathEscape(hash)) {
				stream.Commit(false)
			} else {
				stream.Commit(true)
			}
			// 正式添加该文件
			e = es.AddVersion(stream.Name, stream.Hash, stream.Location, stream.Size)
			if e != nil {
				errmsg.ErrLog(c, http.StatusInternalServerError, e.Error())
			}
			return
		}
	}
}
