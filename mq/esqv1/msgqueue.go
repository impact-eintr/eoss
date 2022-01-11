package esqv1

import (
	"log"
	"strings"
	"sync"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

var (
	TOPIC_heartbeat = "心跳"
	TOPIC_filereq   = "文件定位请求"
)

const (
	ESQV1_ADDRESS = "172.18.0.3:6430"
)

var (
	FileTimeMap = make(map[string]chan string)
	Locker      sync.RWMutex
)

type LocateRouter struct {
	enet.BaseRouter
}

func (this *LocateRouter) Handle(req iface.IRequest) {
	log.Println(string(req.GetData()))
	msgs := strings.Split(string(req.GetData()), "\t") // dataNode的通信地址\t文件名\n时间戳
	Locker.Lock()
	if ch, ok := FileTimeMap[msgs[1]]; ok {
		ch <- msgs[0] // TODO 这个channel关闭了怎么办
	}
	Locker.Unlock()
}

// 这个函数给ApiNode用的
func Test() {
	s := enet.NewServer("tcp4")
	// 添加路由
	router := &LocateRouter{}
	s.AddRouter(20, router)

	s.Serve()
}
