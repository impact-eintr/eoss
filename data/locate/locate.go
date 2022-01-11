package locate

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
)

func Locate(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func StartLocate() {
	if os.Getenv("RABBITMQ_SERVER") != "" {
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		defer q.Close()
		q.Bind("dataServers")
		c := q.Consume()
		for msg := range c {
			object, e := strconv.Unquote(string(msg.Body))
			if e != nil {
				panic(e)
			}
			if Locate("/tmp/objects/" + object) {
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
				object := s[1]
				timeStamp := s[2]

				// 先检查有没有文件
				if Locate("/tmp/eoss/objects/" + object) {
					fmt.Println("即将通知addr:", apiAddr)

					c, err := net.Dial("tcp4", apiAddr)
					if err != nil {
						fmt.Println("client start err ", err)
						return
					}
					// 消息内容 ip:port\t文件名\n时间戳
					addr, _ := dp.Pack(enet.NewMsgPackage(20,
						[]byte(os.Getenv("LISTEN_ADDRESS")+":"+os.Getenv("LISTEN_PORT")+"\t"+object+"\n"+timeStamp)))
					_, err = c.Write(addr)
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
