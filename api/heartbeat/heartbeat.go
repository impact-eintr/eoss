package heartbeat

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

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

func ChooseRandomDataServers(n int, exclude map[int]string) (ds []string) {
	candidates := make([]string, 0)
	reverseExcludeMap := make(map[string]int)
	for id, addr := range exclude {
		reverseExcludeMap[addr] = id
	}
	servers := GetDataServers()
	for i := range servers {
		s := servers[i]
		_, excluded := reverseExcludeMap[s]
		if !excluded {
			candidates = append(candidates, s)
		}
	}
	length := len(candidates)
	if length < n {
		return
	}
	p := rand.Perm(length)
	for i := 0; i < n; i++ {
		ds = append(ds, candidates[p[i]])
	}
	return
}

func GetDataServers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	ds := make([]string, 0)
	for s := range dataServers {
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
			//cli := esqv1.ChooseQueueInCluster("127.0.0.1:2379")
			cli := esqv1.ChooseQueue(os.Getenv("ESQ_SERVER"))
			cli.Config(esqv1.TOPIC_heartbeat, 1, 2, 5, 3) // 配置为自动回复 广播 5s失效 重试3次
			cli.Declare(esqv1.TOPIC_heartbeat, "client"+os.Getenv("LISTEN_ADDRESS"))

			for {
				msg, err := cli.Pop(esqv1.TOPIC_heartbeat, "client"+os.Getenv("LISTEN_ADDRESS"))
				if err != nil {
					log.Println(err)
					break
				}
				mutex.Lock()
				dataServers[msg.Body] = time.Now()
				mutex.Unlock()
			}
		}
	} else {
		panic("需要消息队列!")
	}
}
