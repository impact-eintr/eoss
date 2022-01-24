package locate

import (
	"fmt"
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
	"github.com/impact-eintr/esq"
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
			conn, err := net.Dial("tcp4", os.Getenv("ESQ_SERVER")) // ESQ_SERVER
			if err != nil {
				fmt.Println("client start err ", err)
				time.Sleep(time.Second)
				continue
			}
			defer conn.Close()

			//发封包message消息 只写一次
			msg := esq.PackageProtocol(0, "SUB", esqv1.TOPIC_filereq, os.Getenv("LISTEN_ADDRESS"),
				"有人要戒色？")
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println("write error err ", err)
				return
			}

			// 不停地读
			for {
				data, err := esq.ReadOnce(conn)
				if err != nil {
					fmt.Println(err)
					break
				}
				// 向ApiNode通信
				s := strings.Split(string(data), "\n") // apiIP:api:PORT\n文件名\n时间戳
				apiAddr := s[0]
				hash := s[1]
				timeStamp := s[2]

				// 先检查有没有文件
				ID := Locate(hash) // 文件分片ID
				if ID != -1 {
					c, err := net.Dial("tcp4", apiAddr)
					if err != nil {
						fmt.Println("client start err ", err)
						return
					}
					// 发送的消息内容 ip:port\t文件名\n时间戳-ID
					s := fmt.Sprintf("%s:%s\t%s\n%s-%d", os.Getenv("LISTEN_ADDRESS"),
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
