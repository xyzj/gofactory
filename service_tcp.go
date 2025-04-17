package gofactory

import "github.com/xyzj/toolbox/tcpfactory"

func (s *Service) TcpWrite(target string, msgs ...*tcpfactory.SendMessage) {
	s.tcpserver.WriteTo(target, msgs...)
}
