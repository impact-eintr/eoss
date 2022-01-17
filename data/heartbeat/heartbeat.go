package heartbeat

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
	"github.com/impact-eintr/esq"
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
		for {
			conn, err := net.Dial("tcp4", os.Getenv("ESQ_SERVER")) // ESQ_SERVER
			if err != nil {
				fmt.Println("client start err ", err)
				time.Sleep(time.Second)
				continue
			}
			defer conn.Close()

			for {
				msg := esq.PackageProtocol(0, "PUB", esqv1.TOPIC_heartbeat, os.Getenv("LISTEN_ADDRESS"),
					os.Getenv("LISTEN_ADDRESS")+":"+os.Getenv("LISTEN_PORT"))
				//发封包message消息
				_, err = conn.Write(msg)
				if err != nil {
					fmt.Println("write error err ", err)
					return
				}
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		panic("需要消息队列!")
	}

}
