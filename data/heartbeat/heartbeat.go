package heartbeat

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
)

func StartHeartbeat() {

	if os.Getenv("RABBITMQ_SERVER") != "" {
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		defer q.Close()
		for {
			q.Publish("apiServers", os.Getenv("LISTEN_ADDRESS"))
			time.Sleep(5 * time.Second)
		}

	} else if os.Getenv("ESQ_SERVER") != "" {
		conn, err := net.Dial("tcp4", os.Getenv("ESQ_SERVER"))
		if err != nil {
			fmt.Println("client start err ", err)
			return
		}
		defer conn.Close()

		for {
			dp := enet.GetDataPack()
			msg, _ := dp.Pack(enet.NewMsgPackage(0,
				[]byte(esqv1.TOPIC_heartbeat+"\t"+os.Getenv("LISTEN_ADDRESS")+":"+os.Getenv("LISTEN_PORT"))))
			//发封包message消息
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println("write error err ", err)
				return
			}
			time.Sleep(5 * time.Second)
		}
	} else {
		panic("需要消息队列!")
	}

}
