package heartbeat

import (
	"os"
	"time"

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
		for {
			//cli := esqv1.ChooseQueueInCluster("127.0.0.1:2379")
			cli := esqv1.ChooseQueue(os.Getenv("ESQ_SERVER"))
			cli.Config("heartbeat", 1, 2, 5, 3)
			// 发送消息
			for {
				cli.Push(os.Getenv("LISTEN_ADDRESS")+":"+os.Getenv("LISTEN_PORT"), esqv1.TOPIC_heartbeat, "client*", 0)
				time.Sleep(5 * time.Second)
			}
		}
	} else {
		panic("需要消息队列!")
	}

}
