package gofactory

import (
	"errors"
	"time"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/tcpfactory"
)

type tcpSvr struct {
	helloMessages []*tcpfactory.SendMessage
	client        tcpfactory.Client
	readTimeout   time.Duration
	bind          string
	enable        bool
}

func (opt *tcpSvr) build(l logger.Logger) (*tcpfactory.TCPManager, error) {
	if !opt.enable {
		return nil, errors.New("[tcp] server not enable")
	}
	tcpserver, err := tcpfactory.NewTcpFactory(
		tcpfactory.OptBindAddr(opt.bind),
		tcpfactory.OptLogger(l),
		tcpfactory.OptHelloMessages(opt.helloMessages...),
		tcpfactory.OptMatchMultiTargets(true),
		tcpfactory.OptReadTimeout(opt.readTimeout),
		tcpfactory.OptTcpClient(opt.client),
	)
	if err != nil {
		return nil, errors.New("[tcp] build server error:" + err.Error())
		// return nil, err
	}
	return tcpserver, nil
}

var defaultTCPServer = tcpSvr{
	bind:          ":6881",
	client:        &tcpfactory.EchoClient{},
	helloMessages: make([]*tcpfactory.SendMessage, 0),
	readTimeout:   time.Second * 30,
}

type tcpOpts func(o *tcpSvr)

func OptTCPBindAddr(s string) tcpOpts {
	return func(o *tcpSvr) {
		o.bind = s
	}
}

func OptTCPReadTimeout(t time.Duration) tcpOpts {
	return func(o *tcpSvr) {
		o.readTimeout = min(max(t, time.Second), time.Hour)
	}
}

func OptTCPHelloMessages(m []*tcpfactory.SendMessage) tcpOpts {
	return func(o *tcpSvr) {
		o.helloMessages = m
	}
}

func OptTCPClient(c tcpfactory.Client) tcpOpts {
	return func(o *tcpSvr) {
		o.client = c
	}
}
