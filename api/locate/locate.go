package locate

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
)

func Get(ctx *gin.Context) {
	name := ctx.Param("name")
	address := Locate(name)
	if len(address) == 0 {
		errmsg.ErrLog(ctx, http.StatusNotFound, "no any dataNode online")
		return
	}

	ctx.JSON(http.StatusOK, address)

}

func Locate(name string) string {
	if os.Getenv("RABBITMQ_SERVER") != "" {
		// rabbitmq
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		q.Publish("dataServers", name)
		c := q.Consume()
		go func() {
			time.Sleep(time.Second)
			q.Close()
		}()
		msg := <-c
		s, _ := strconv.Unquote(string(msg.Body))
		return s

	} else if os.Getenv("ESQ_SERVER") != "" {
		// esqv1
		conn, err := net.Dial("tcp4", os.Getenv("ESQ_SERVER")) // ESQ_SERVER
		if err != nil {
			fmt.Println("client start err ", err)
			return ""
		}
		defer conn.Close()

		mapKey := fmt.Sprintf("%s\n%d", name, time.Now().Unix()) // 向dataNode传递的消息 ip:port\nName\nTime

		// 向dataNode集群广播消息
		dp := enet.GetDataPack()
		localServer := fmt.Sprintf("%s:%d", os.Getenv("LISTEN_ADDRESS"), enet.GlobalObject.Port)
		msg := fmt.Sprintf("%s\t%s\n%s", esqv1.TOPIC_filereq, localServer, mapKey)
		data, _ := dp.Pack(enet.NewMsgPackage(0, []byte(msg)))
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("write error err ", err)
			return ""
		}

		// 等待定位结果
		// 注册消息
		ch := make(chan string)
		esqv1.Locker.Lock()
		esqv1.FileTimeMap[mapKey] = ch // TODO 无缓冲
		esqv1.Locker.Unlock()

		// 注销消息
		defer func() {
			esqv1.Locker.Lock()
			delete(esqv1.FileTimeMap, mapKey)
			esqv1.Locker.Unlock()
		}()

		timer := time.NewTimer(1 * time.Second)
		select {
		case s := <-ch: // 最终结果
			// TODO 小心这里的 close 坑
			return s
		case <-timer.C:
			return ""
		}

	} else {
		return ""
	}
}

func Exist(name string) bool {
	return Locate(name) != ""
}
