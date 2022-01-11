package main

import (
	"fmt"
	"strings"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
	"github.com/impact-eintr/esq"
	"github.com/impact-eintr/esq/mq"
)

type PubRouter struct {
	enet.BaseRouter
}

func (this *PubRouter) Handle(req iface.IRequest) {
	msgs := strings.Split(string(req.GetData()), "\t")
	esq.Pub(msgs[0], msgs[1], m, nil)
}

type SubRouter struct {
	enet.BaseRouter
}

func (this *SubRouter) Handle(req iface.IRequest) {
	msgs := strings.Split(string(req.GetData()), "\t")
	go esq.Sub(msgs[0], m, func(v interface{}, ch chan bool) {
		err := req.GetConnection().SendTcpMsg(2020, []byte(v.(string)))
		if err != nil {
			ch <- true
			fmt.Println("退出原因：", err)
		}
	})
}

var m = mq.NewClient()

func main() {
	m.SetConditions(10)
	s := enet.NewServer("tcp4")
	// 添加路由
	s.AddRouter(0, &PubRouter{})  // Publish
	s.AddRouter(10, &SubRouter{}) // Subscribe

	s.Serve()
}
