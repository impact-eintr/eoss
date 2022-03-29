package esqv1

import (
	"strings"
	"sync"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

var (
	TOPIC_heartbeat = "Ping..."
	TOPIC_filereq   = "File..."
	TOPIC_fileresp  = "Found..."
)

var (
	FileTimeMap = make(map[string]chan string)
	Locker      sync.RWMutex
)

type LocateRouter struct {
	enet.BaseRouter
}

func (this *LocateRouter) Handle(req iface.IRequest) {
	// 收到的消息内容 ip:port\t文件名-时间戳-ID
	addr := strings.Split(string(req.GetData()), "\t")[0]
	name := strings.Split(strings.Split(string(req.GetData()), "\t")[1], "-")[0]
	stamp := strings.Split(strings.Split(string(req.GetData()), "\t")[1], "-")[1]
	id := strings.Split(string(req.GetData()), "-")[2]
	mapKey := name + "-" + stamp

	Locker.Lock()
	if ch, ok := FileTimeMap[mapKey]; ok {
		ch <- addr + "-" + id // TODO 这个channel关闭了怎么办
	}
	Locker.Unlock()
}

// 这个函数给ApiNode用的
func ListenFileResponse() {
	s := enet.NewServer("tcp4")
	// 添加路由
	router := &LocateRouter{}
	s.AddRouter(20, router)

	s.Serve()
}
