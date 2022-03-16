package heartbeat

import (
	"log"
	"os"
	"time"

	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
)

func StartHeartbeat() {
	if os.Getenv("RAFTD_SERVER") != "" {
		for {
			// 发送消息
			cli := esqv1.ChooseQueueInCluster(os.Getenv("RAFTD_SERVER"))
			cli.Config("heartbeat", 1, 2, 5, 3)
			err := cli.Push(os.Getenv("LISTEN_ADDRESS")+":"+os.Getenv("LISTEN_PORT"), esqv1.TOPIC_heartbeat, "client*", 0)
			if err != nil {
				log.Println(err)
				break
			}
			time.Sleep(5 * time.Second)
		}
	} else if os.Getenv("ESQ_SERVER") != "" {
		for {
			cli := esqv1.ChooseQueue(os.Getenv("ESQ_SERVER"))
			cli.Config("heartbeat", 1, 2, 5, 3)
			// 发送消息
			for {
				err := cli.Push(os.Getenv("LISTEN_ADDRESS")+":"+os.Getenv("LISTEN_PORT"), esqv1.TOPIC_heartbeat, "client*", 0)
				if err != nil {
					log.Println(err)
					break
				}
				time.Sleep(5 * time.Second)
			}
		}
	} else if os.Getenv("RABBITMQ_SERVER") != "" {
		q := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
		defer q.Close()
		for {
			q.Publish("apiServers", os.Getenv("LISTEN_ADDRESS"))
			time.Sleep(5 * time.Second)
		}
	} else {
		panic("需要消息队列!")
	}

}
