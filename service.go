package gofactory

import (
	"crypto/tls"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xyzj/mqtt-server/cmd/server"
	"github.com/xyzj/toolbox/db"
	"github.com/xyzj/toolbox/httpclient"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
	"github.com/xyzj/toolbox/tcpfactory"
)

type Service struct {
	opt        *Opt
	httpcli    *httpclient.Client
	boltcli    *db.BoltDB
	tcpserver  *tcpfactory.TCPManager
	mqttbroker *server.MqttServer
	webserver  map[string]*http.Server
}

func New(opts ...Opts) (*Service, error) {
	opt := Opt{
		logg: logger.NewConsoleLogger(),
		mode: Debug,
	}
	for _, o := range opts {
		o(&opt)
	}
	s := &Service{
		opt: &opt,
		httpcli: httpclient.New(
			httpclient.OptLogger(opt.logg),
			httpclient.OptTLS(&tls.Config{InsecureSkipVerify: true}),
		),
	}
	if opt.mode == Debug {
		opt.logg.SetLevel(logger.LogDebug)
	}
	var err error
	// services
	// tcp
	if opt.tcpServer.enable {
		s.tcpserver, err = opt.tcpServer.build(opt.logg)
		if err != nil {
			opt.logg.Error("build tcp server error:" + err.Error())
		}
	}
	// mqtt broker
	if opt.mqttBroker.enable {
		s.mqttbroker, err = opt.mqttBroker.build(opt.logg, opt.mode)
		if err != nil {
			opt.logg.Error("build mqtt server error:" + err.Error())
		}
	}
	// web
	if opt.webServer.enable {
		s.webserver, err = opt.webServer.build(opt.logg)
		if err != nil {
			opt.logg.Error("build web server error:" + err.Error())
		}
	}
	// clients
	// boltdb
	if s.opt.boltname != "" {
		s.boltcli, err = db.NewBolt(opt.boltname)
		if err != nil {
			opt.logg.Error("create or load boltdb error:" + err.Error())
		} else {
			p, err := filepath.Abs(opt.boltname)
			if err != nil {
				p = opt.boltname
			}
			opt.logg.System("[bolt] create or load boltdb from:" + p)
		}
	}
	return s, nil
}

func (s *Service) Start() {
	loopfunc.GoFunc(func(params ...any) {
		s.Run()
	}, "service", s.opt.logg.DefaultWriter())
}

func (s *Service) Run() {
	wg := sync.WaitGroup{}
	if s.tcpserver != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.tcpserver.Listen()
			if err != nil {
				s.opt.logg.Error("[tcp] listen error:" + err.Error())
				return
			}
		}()
	}
	if s.mqttbroker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.mqttbroker.Run()
		}()
	}
	if s.webserver != nil {
		h, err := s.opt.webServer.buildRoutes()
		if err != nil {
			s.opt.logg.Error("[web] build routes error:" + err.Error())
			return
		}
		wg.Add(2)
		go func() {
			defer wg.Done()
			if ss, ok := s.webserver["https"]; ok {
				ss.Handler = h
				s.opt.logg.System("[web] https listening to " + ss.Addr)
				err = ss.ListenAndServeTLS("", "")
				if err != nil {
					s.opt.logg.Error("[web] serve tls web service failed:" + err.Error())
				}
			}
		}()
		go func() {
			defer wg.Done()
			if ss, ok := s.webserver["http"]; ok {
				ss.Handler = h
				s.opt.logg.System("[web] http listening to " + ss.Addr)
				err = ss.ListenAndServe()
				if err != nil {
					s.opt.logg.Error("[web] serve web service failed:" + err.Error())
				}
			}
		}()
	}
	// clients
	// discover
	if s.opt.discover.enable {
		if len(s.opt.discover.svrInfo.RegisterAddress) == 0 {
		}
		err := s.opt.discover.build(s.opt.logg)
		if err != nil {
			s.opt.logg.Error("build discover error:" + err.Error())
		}
	}
	// redis
	if s.opt.cliredis.enable {
		err := s.opt.cliredis.build(s.opt.logg)
		if err != nil {
			s.opt.logg.Error("build redis client error:" + err.Error())
		}
	}
	// mqtt
	if s.opt.climqtt.enable {
		err := s.opt.climqtt.build(s.opt.logg)
		if err != nil {
			s.opt.logg.Error("build mqtt client error:" + err.Error())
		}
	}
	// rmq
	if s.opt.clirmq.enable {
		err := s.opt.clirmq.build(s.opt.logg)
		if err != nil {
			s.opt.logg.Error("build rmq clients error:" + err.Error())
		}
	}
	wg.Wait()
}

func (s *Service) AppendRootPath(ss, sep string) string {
	if !s.opt.discover.enable {
		return ss
	}
	if s.opt.discover.svrInfo.RootPath == "" {
		return ss
	}
	if strings.HasPrefix(ss, s.opt.discover.svrInfo.RootPath) {
		return ss
	}
	if sep != "" && strings.HasPrefix(ss, sep) {
		return s.opt.discover.svrInfo.RootPath + ss
	}
	return s.opt.discover.svrInfo.RootPath + sep + ss
}
