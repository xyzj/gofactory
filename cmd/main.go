package main

import (
	"github.com/xyzj/tcpfactory"
	"github.com/xyzj/toolbox/logger"
)

func main() {
	m, err := tcpfactory.NewTcpFactory(":6819",
		tcpfactory.OptLogger(logger.NewConsoleLogger()),
		tcpfactory.OptTcpFactory(&tcpfactory.EchoClient{}),
	)
	if err != nil {
		panic(err)
	}
	m.Listen()
}
