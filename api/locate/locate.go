package locate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/eoss/errmsg"
	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
	"github.com/impact-eintr/eoss/rs"
	"github.com/impact-eintr/eoss/types"
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

func Locate(name string) (locateInfo map[int]string) {
	locateInfo = make(map[int]string)
	// 根据集群配置选择不同的通信
	if os.Getenv("RAFTD_SERVER") != "" {
		cli := esqv1.ChooseQueueInCluster(os.Getenv("RAFTD_SERVER"))
		//cli.Config(esqv1.TOPIC_filereq, 0, 2, 5, 3) // 不自动回复了

		// 向dataNode集群广播消息:  文件名-ip:时间戳
		bindKey := fmt.Sprintf("%s-%s:%d", name, os.Getenv("LISTEN_ADDRESS"), time.Now().Unix())
		cli.Push(bindKey, esqv1.TOPIC_filereq, "client*", 0)
		bindKey = url.QueryEscape(bindKey)

		// 等待定位结果
		cli = esqv1.ChooseQueueInCluster(os.Getenv("RAFTD_SERVER"))
		cli.Declare(esqv1.TOPIC_fileresp, bindKey)
		cli.Config(esqv1.TOPIC_fileresp, 1, 1, 5, 3)

		timer := time.NewTimer(2 * time.Second)
		ch := make(chan string, rs.ALL_SHARDS)
		go func() {
			for {
				msg, err := cli.Pop(esqv1.TOPIC_fileresp, bindKey)
				if err != nil {
					return // TODO 处理一下
				}
				select {
				case <-ch:
					log.Println("订阅协程退出")
					return
				case ch <- msg.Body:
				}
			}
		}()
		// 取够 ALL_SHARDS 次 或者 超时
		for i := 0; i < rs.ALL_SHARDS; i++ {
			select {
			case <-timer.C:
				close(ch)
				return
			case msg := <-ch:
				// 解码 msg
				s := strings.Split(msg, "-")
				addr := s[0]
				id, _ := strconv.Atoi(s[1])
				locateInfo[id] = addr
			}
		}
		return
	} else if os.Getenv("ESQ_SERVER") != "" {
		cli := esqv1.ChooseQueue(os.Getenv("ESQ_SERVER"))
		cli.Config(esqv1.TOPIC_filereq, 0, 2, 5, 3) // 不自动回复了

		mapKey := fmt.Sprintf("%s-%d", name, time.Now().Unix())

		// 向dataNode集群广播消息:  ip:port-文件名-时间戳
		localServer := fmt.Sprintf("%s:%d", os.Getenv("LISTEN_ADDRESS"), enet.GlobalObject.Port)
		cli.Push(fmt.Sprintf("%s-%s", localServer, mapKey), esqv1.TOPIC_filereq, "client*", 0)

		// 等待定位结果
		// 注册消息
		ch := make(chan string, rs.ALL_SHARDS)
		esqv1.Locker.Lock()
		esqv1.FileTimeMap[mapKey] = ch
		esqv1.Locker.Unlock()

		// 注销消息
		defer func() {
			esqv1.Locker.Lock()
			delete(esqv1.FileTimeMap, mapKey)
			esqv1.Locker.Unlock()
		}()

		// 取够 ALL_SHARDS 次 或者 超时
		timer := time.NewTimer(1 * time.Second)
		for i := 0; i < rs.ALL_SHARDS; i++ {
			select {
			// TODO 小心这里的 close 坑
			case addr_id := <-ch: // 最终结果 ip:port-id
				s := strings.Split(addr_id, "-")
				addr := s[0]
				id, _ := strconv.Atoi(s[1])
				locateInfo[id] = addr
			case <-timer.C:
				return
			}
		}
		return
	} else if os.Getenv("RABBITMQ_SERVER") != "" {
		// rabbitmq
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		q.Publish("dataServers", name)
		c := q.Consume()
		go func() {
			time.Sleep(time.Second)
			q.Close()
		}()
		for i := 0; i < rs.ALL_SHARDS; i++ {
			msg := <-c
			if len(msg.Body) == 0 {
				return
			}
			var info types.LocateMessage
			json.Unmarshal(msg.Body, &info)
			locateInfo[info.Id] = info.Addr
		}
		return
	} else {
		return
	}
}

func Exist(name string) bool {
	return len(Locate(name)) >= rs.DATA_SHARDS
}
