package heartbeat

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
)

var dataServers = make(map[string]time.Time)
var mutex sync.Mutex

func removeExpiredDataServer() {
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		for s, t := range dataServers {
			if t.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, s)
			}
		}
		mutex.Unlock()
	}
}

func ChooseRandomDataServer() string {
	ds := GetDataServers()
	n := len(ds)
	if n == 0 {
		return ""
	}
	return ds[rand.Intn(n)]
}

func GetDataServers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	ds := make([]string, 0)
	for s, _ := range dataServers {
		ds = append(ds, s)
	}
	fmt.Println("数据服务器名单 ", ds)
	return ds
}

func ListenHeartbeat() {
	if os.Getenv("RABBITMQ_SERVER") != "" {
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		defer q.Close()
		q.Bind("apiServers")
		c := q.Consume()
		go removeExpiredDataServer()
		for msg := range c {
			dataServer, e := strconv.Unquote(string(msg.Body))
			if e != nil {
				panic(e)
			}
			mutex.Lock()
			dataServers[dataServer] = time.Now()
			mutex.Unlock()
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
			dp := enet.GetDataPack()
			msg, _ := dp.Pack(enet.NewMsgPackage(10, []byte(esqv1.TOPIC_heartbeat+"\t"+"我经常帮助一些翘家的人")))
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println("write error err ", err)
				continue
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
					break
				}

				//msg 是有data数据的，需要再次读取data数据
				if msgHead.GetDataLen() > 0 {
					msg := msgHead.(*enet.Message)
					msg.Data = make([]byte, msg.GetDataLen())

					//根据dataLen从io中读取字节流
					_, err := io.ReadFull(conn, msg.Data)
					if err != nil {
						fmt.Println("server unpack data err:", err)
						break
					}

					mutex.Lock()
					dataServers[string(msg.Data)] = time.Now()
					mutex.Unlock()
				}
			}
		}
	} else {
		panic("需要消息队列!")
	}
}
