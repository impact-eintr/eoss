package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/judwhite/go-svc"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
	"github.com/impact-eintr/esq"
	"github.com/impact-eintr/esq/mq"
)

type Router struct {
	enet.BaseRouter
}

func (this *Router) Handle(req iface.IRequest) {

	data := req.GetData()
	switch esq.ParseCommand(data) {
	case "PUB":
		esq.Pub(esq.ParseTopic(data), esq.ParseMsg(data), m, nil)
	case "SUB":
		go esq.Sub(esq.ParseTopic(data), esq.ParseSrcHost(data), m, func(v interface{}, ch chan bool) {
			err := req.GetConnection().SendTcpMsg(2020, v.([]byte))
			if err != nil {
				ch <- true
			}
		})
	default:
		log.Println("invalid command type", esq.ParseCommand(data))
	}
}

var (
	m     = new(mq.Client)
	usage = "[NOTICE] If you need to enable debugging,please set the environment variable through `export esq_debug`"
)

type program struct {
	svr iface.IServer
}

func (p *program) Init(env svc.Environment) error {
	p.svr.AddRouter(0, &Router{})
	return nil
}

func (p *program) Start() error {
	go p.svr.Start()
	return nil
}

func (p *program) Stop() error {
	m.Close()
	p.svr.Stop()
	return nil
}

func init() {
	path := flag.String("path", "/tmp/esq", "queue path")
	filesize := flag.Int64("filesize", 65536, "file size")
	minsize := flag.Int("minsize", 0, "min msg size")
	maxsize := flag.Int("maxsize", math.MaxInt32, "max msg size")
	sync := flag.Int64("sync", 1024, "sync count")

	flag.Parse()

	m = mq.NewClient(
		*path,
		*filesize,
		int32(*minsize),
		int32(*maxsize),
		*sync,
		time.Second, // 同步计时
	)
}

func main() {
	fmt.Println(usage)

	prg := program{
		svr: enet.NewServer("tcp4"),
	}

	if err := svc.Run(&prg); err != nil {
		log.Fatal(err)
	}
}
