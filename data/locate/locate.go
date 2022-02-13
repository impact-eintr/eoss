package locate

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
)

var objects = make(map[string]int)
var mutex sync.Mutex

func Locate(hash string) int {
	mutex.Lock()
	id, ok := objects[hash]
	mutex.Unlock()
	if !ok {
		return -1
	}
	return id
}

func Add(hash string, id int) {
	mutex.Lock()
	objects[hash] = id
	mutex.Unlock()
}

func Del(hash string) {
	mutex.Lock()
	delete(objects, hash)
	mutex.Unlock()
}

func StartLocate() {
	if os.Getenv("RABBITMQ_SERVER") != "" {
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		defer q.Close()
		q.Bind("dataServers")
		c := q.Consume()
		for msg := range c {
			hash, e := strconv.Unquote(string(msg.Body))
			if e != nil {
				panic(e)
			}
			if Locate(hash) != -1 {
				q.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS"))
			}
		}
	} else if os.Getenv("ESQ_SERVER") != "" {
		for {
			//cli := esqv1.ChooseQueueInCluster("127.0.0.1:2379")
			cli := esqv1.ChooseQueue(os.Getenv("ESQ_SERVER"))
			cli.Config(esqv1.TOPIC_filereq, 0, 2, 5, 3) // 不要自动回复
			cli.Declare(esqv1.TOPIC_filereq, "client"+os.Getenv("LISTEN_ADDRESS"))

			// 发送消息
			func() {
				for {
					msg, err := cli.Pop(esqv1.TOPIC_filereq, "client"+os.Getenv("LISTEN_ADDRESS"))
					if err != nil {
						defer time.Sleep(100 * time.Millisecond)
						break
					}
					// 回复消息
					err = cli.Ack(esqv1.TOPIC_filereq, "client"+os.Getenv("LISTEN_ADDRESS"), msg.Id)
					if err != nil {
						defer time.Sleep(100 * time.Millisecond)
						break
					}

					// 向ApiNode通信
					s := strings.Split(msg.Body, "-") // ip:port-文件名-时间戳
					apiAddr := s[0]
					hash := s[1]
					timeStamp := s[2]

					// 先检查有没有文件
					ID := Locate(hash) // 文件分片ID
					if ID != -1 {
						log.Println("找到了文件", hash, ID)
						c, err := net.Dial("tcp4", apiAddr)
						if err != nil {
							fmt.Println("client start err ", err)
							return
						}
						// 发送的消息内容 ip:port\t文件名-时间戳-ID
						s := fmt.Sprintf("%s:%s\t%s-%s-%d", os.Getenv("LISTEN_ADDRESS"),
							os.Getenv("LISTEN_PORT"), hash, timeStamp, ID)
						dp := enet.GetDataPack()
						respMsg, _ := dp.Pack(enet.NewMsgPackage(20,
							[]byte(s)))
						_, err = c.Write(respMsg)
						if err != nil {
							fmt.Println("write error err ", err)
							return
						}
						c.Close()
					}
				}
			}()
		}
	} else {
		panic("需要消息队列")
	}

}

// 对外提供服务前先要知道自己有哪些文件
func CollectObjects() {
	files, _ := filepath.Glob("/tmp/eoss/objects/*")
	for i := range files {
		file := strings.Split(filepath.Base(files[i]), ".")
		if len(file) != 3 {
			panic(files[i])
		}
		hash := file[0]
		id, e := strconv.Atoi(file[1])
		if e != nil {
			panic(e)
		}
		objects[hash] = id
	}
}
