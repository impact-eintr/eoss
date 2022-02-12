package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

	"github.com/impact-eintr/esq/gnode"
	"github.com/kardianos/service"
)

var logger service.Logger

type program struct {
	gn   *gnode.Gnode
	once sync.Once
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	cfg := gnode.NewGnodeConfig()
	p.gn = gnode.New(cfg)

	go func() {
		p.gn.Run()
	}()

	// listen pprof
	go func() {
		if err := http.ListenAndServe("0.0.0.0:8877", nil); err != nil {
			log.Println(err)
		}
	}()

	return nil

}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	p.once.Do(func() {
		p.gn.Exit()
	})
	return nil

}

// TODO 如果启动集群模式 对于当前场景而言 必须使用 权重模式 来投送消息
//      esq1:1 esq2:2 esq3:3 默认会往最新上线的节点发送消息
//      这样所有 apiServer dataServer 的信息就能同步到一个节点上处理
//      如果这个esq节点挂了 就从存活的esq中找最新的一个发送消息
//      这样 现在的esq集群是一个 会丢失消息 的 高可用集群

func main() {
	svcConfig := &service.Config{
		Name:        "esq-node",
		DisplayName: "esq-node",
		Description: "This is an esq-node service.",
		Arguments:   []string{"-config_file="},
		Option:      make(map[string]interface{}),
	}

	svcConfig.Option["LogOutput"] = true

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		var e error
		switch os.Args[1] {
		case "install":
			if err := s.Install(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service install success!")
			}
			if err := s.Start(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service start success!")
			}
		case "uninstall":
			if err := s.Stop(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service stop success!")
			}
			if err := s.Uninstall(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service uninstall success!")
			}
		case "start":
			if err := s.Start(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service start success!")
			}
		case "stop":
			if err := s.Stop(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service stop success!")
			}
		case "restart":
			if err := s.Stop(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service stop success!")
			}
			if err := s.Start(); err != nil {
				e = logger.Error(err)
			} else {
				e = logger.Info("gnode.service start success!")
			}
		case "status":
			if status, err := s.Status(); err != nil {
				e = logger.Error(err)
			} else {
				if int(status) == 1 {
					e = logger.Info("gnode.service is running.")
				} else {
					e = logger.Info("gnode.service is stop.")
				}
			}
		default:
			goto Run
		}

		if e != nil {
			log.Fatalln(e)
		}
	}
Run:
	if err = s.Run(); err != nil {
		log.Fatalln(err)
	}
}
