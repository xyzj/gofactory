package gofactory

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xyzj/toolbox/logger"
)

type RunMode byte

const (
	Dev RunMode = iota
	Debug
	Release
)

type emptySvr struct {
	enable bool
}

type Opt struct {
	// services
	tcpServer   tcpSvr
	mqttBroker  mqttBroker
	webServer   webSvr
	emptyServer emptySvr
	// clients
	discover discover
	cliredis cliRedis
	climqtt  cliMqtt
	clirmq   cliRmq
	clidb    cliDB
	boltname string
	// base config
	logg logger.Logger
	mode RunMode
}

type Opts func(opt *Opt)

func checkTCPAddr(s string) (*net.TCPAddr, bool) {
	if s == "" {
		return nil, false
	}
	if a, err := net.ResolveTCPAddr("tcp", s); err != nil || a.Port == 0 {
		return nil, false
	} else {
		return a, true
	}
}

func SetMode(m RunMode) Opts {
	return func(o *Opt) {
		o.mode = m
		if m == Release {
			gin.SetMode(gin.ReleaseMode)
		}
	}
}

func WithLogger(l logger.Logger) Opts {
	return func(o *Opt) {
		o.logg = l
	}
}

func WithTCPServer(opts ...tcpOpts) Opts {
	return func(o *Opt) {
		o.tcpServer = defaultTCPServer
		o.tcpServer.enable = true
		for _, v := range opts {
			v(&o.tcpServer)
		}
	}
}

func WithMQTTBroker(opts ...mqttBrokerOpts) Opts {
	return func(o *Opt) {
		o.mqttBroker = defaultMqttBroker
		o.mqttBroker.enable = true
		for _, v := range opts {
			v(&o.mqttBroker)
		}
	}
}

func WithWebServer(opts ...webOpts) Opts {
	return func(o *Opt) {
		o.webServer = defaultWebServer
		o.webServer.enable = true
		for _, v := range opts {
			v(&o.webServer)
		}
	}
}

func WithEmptyServer() Opts {
	return func(o *Opt) {
		o.emptyServer = emptySvr{
			enable: true,
		}
	}
}

func WithRedisClient(opts ...redisOpts) Opts {
	return func(o *Opt) {
		o.cliredis = cliRedis{
			user:         "",
			pwd:          "",
			addr:         "127.0.0.1:6379",
			database:     0,
			loaded:       atomic.Bool{},
			readTimeout:  time.Second * 5,
			writeTimeout: time.Second * 10,
		}
		o.cliredis.enable = true
		for _, v := range opts {
			v(&o.cliredis)
		}
	}
}

func WithDiscover(opts ...discoverOpts) Opts {
	return func(o *Opt) {
		o.discover = discover{
			svrInfo:         &svrinfo{},
			infoTimeout:     time.Second * 9,
			publishInterval: time.Second * 4,
			udpPort:         9000,
		}
		o.discover.enable = true
		for _, v := range opts {
			v(&o.discover)
		}
	}
}

func WithMqttClient(opts ...mqttOpts) Opts {
	return func(o *Opt) {
		o.climqtt = cliMqtt{
			addr:               "127.0.0.1:1883",
			user:               "",
			pwd:                "",
			clientID:           "gofactory",
			enable:             true,
			enableFailureCache: true,
			failureCacheMax:    100,
			failureCacheExpire: time.Minute * 5,
		}
		for _, v := range opts {
			v(&o.climqtt)
		}
	}
}

func WithRmqProducer(opts ...rmqOpts) Opts {
	return func(o *Opt) {
		o.clirmq = cliRmq{
			addr:     "127.0.0.1:5672",
			user:     "guest",
			pwd:      "guest",
			vhost:    "/",
			exchange: "",
			enableP:  true,
			enable:   true,
		}
		for _, v := range opts {
			v(&o.clirmq)
		}
	}
}

func WithRMQConsumer(opts ...rmqOpts) Opts {
	return func(o *Opt) {
		o.clirmq = cliRmq{
			addr:            "127.0.0.1:5672",
			user:            "guest",
			pwd:             "guest",
			vhost:           "/",
			exchange:        "",
			queueName:       time.Now().Format("Jan-01-02_15-04-05"),
			queueDurable:    false,
			queueAutoDelete: true,
			enableC:         true,
			enable:          true,
		}
		for _, v := range opts {
			v(&o.clirmq)
		}
	}
}

func WithBoltDB(name string) Opts {
	return func(o *Opt) {
		o.boltname = name
	}
}

func WithDBClient(opts ...dbOpts) Opts {
	return func(o *Opt) {
		o.clidb = cliDB{
			addr:     "127.0.0.1:3306",
			user:     "root",
			pwd:      "root",
			database: []string{"test"},
			enable:   true,
		}
		for _, v := range opts {
			v(&o.clidb)
		}
	}
}
