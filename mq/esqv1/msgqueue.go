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

var (
	FileTimeMap = make(map[string]chan string)
	Locker      sync.RWMutex
)

type LocateRouter struct {
	enet.BaseRouter
}

func (this *LocateRouter) Handle(req iface.IRequest) {
	log.Println(string(req.GetData()))
	// 收到的消息内容 ip:port\t文件名\n时间戳-ID
	addr := strings.Split(string(req.GetData()), "\t")[0]
	mapKey := strings.Split(strings.Split(string(req.GetData()), "\t")[1], "-")[0]
	id := strings.Split(string(req.GetData()), "-")[1]
	log.Println(addr, mapKey, id)
	Locker.Lock()
	if ch, ok := FileTimeMap[mapKey]; ok {
		ch <- addr + "-" + id // TODO 这个channel关闭了怎么办
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
