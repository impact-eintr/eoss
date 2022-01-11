package locate

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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
		conn, err := net.Dial("tcp4", os.Getenv("ESQ_SERVER"))
		if err != nil {
			fmt.Println("client start err ", err)
			return
		}
		defer conn.Close()

		//发封包message消息 只写一次
		dp := enet.GetDataPack()
		msg, _ := dp.Pack(enet.NewMsgPackage(10, []byte(esqv1.TOPIC_filereq+"\t"+"谁想要要色图")))
		_, err = conn.Write(msg)
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

		// 不停地读
		for {
			//先读出流中的head部分
			headData := make([]byte, dp.GetHeadLen())
			_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
			if err != nil {
				fmt.Println("read head error")
				break
			}
			//将headData字节流 拆包到msg中
			msgHead, err := dp.Unpack(headData)
			if err != nil {
				fmt.Println("server unpack err:", err)
				return
			}

			if msgHead.GetDataLen() > 0 {
				//msg 是有data数据的，需要再次读取data数据
				msg := msgHead.(*enet.Message)
				msg.Data = make([]byte, msg.GetDataLen())

				//根据dataLen从io中读取字节流
				_, err := io.ReadFull(conn, msg.Data)
				if err != nil {
					fmt.Println("server unpack data err:", err)
					return
				}

				// 向ApiNode通信
				s := strings.Split(string(msg.GetData()), "\n") // apiIP:api:PORT\n文件名\n时间戳
				apiAddr := s[0]
				hash := s[1]
				timeStamp := s[2]

				// 先检查有没有文件
				ID := Locate(hash) // 文件分片ID
				if ID != -1 {
					fmt.Println("即将通知addr:", apiAddr)

					c, err := net.Dial("tcp4", apiAddr)
					if err != nil {
						fmt.Println("client start err ", err)
						return
					}
					// 发送的消息内容 ip:port\t文件名\n时间戳-ID
					s := fmt.Sprintf("%s:%s\t%s\n%s-%d", os.Getenv("LISTEN_ADDRESS"),
						os.Getenv("LISTEN_PORT"), hash, timeStamp, ID)
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
		log.Println(file)
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
