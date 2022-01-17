package heartbeat

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/impact-eintr/eoss/mq/esqv1"
	"github.com/impact-eintr/eoss/mq/rabbitmq"
	"github.com/impact-eintr/esq"
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
			conn, err := net.Dial("tcp4", os.Getenv("ESQ_SERVER")) // ESQ_SERVER
			if err != nil {
				fmt.Println("client start err ", err)
				time.Sleep(time.Second)
				continue
			}
			defer conn.Close()

			//发封包message消息 只写一次
			msg := esq.PackageProtocol(0, "SUB", esqv1.TOPIC_heartbeat, "1.1.0.1", "我经常帮助一些翘家的人")
			_, err = conn.Write(msg)
			if err != nil {
				fmt.Println("write error err ", err)
				continue
			}

			// 不停地读
			for {
				data, err := esq.ReadOnce(conn)
				if err != nil {
					fmt.Println(err)
					break
				}

				mutex.Lock()
				dataServers[string(data)] = time.Now()
				mutex.Unlock()
			}
		}
	} else {
		panic("需要消息队列!")
	}
}
