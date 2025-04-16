package gofactory

import (
	"errors"
	"net/http"

	"github.com/xyzj/toolbox/httpclient"
)

func (s *Service) PickService(name string, protocol ProtocolType) (string, error) {
	var ss string
	var ok bool
	s.opt.discover.infos.ForEach(func(key string, value svrinfo) bool {
		if value.SvrName == name {
			if ss, ok = value.RegisterAddress[protocol]; ok {
				return false
			}
		}
		return true
	})
	if ok {
		return ss, nil
	}
	return "", errors.New("service not found")
}

func (s *Service) PickAll() map[string]string {
	var ss map[string]string
	s.opt.discover.infos.ForEach(func(key string, value svrinfo) bool {
		ss[key] = value.Json()
		return true
	})
	return ss
}

func (s *Service) DoRequest(req *http.Request, opts ...httpclient.ReqOpts) (int, []byte, map[string]string, error) {
	return s.httpcli.DoRequest(req, opts...)
}
