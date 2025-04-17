package gofactory

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/logger"
)

//go:embed favicon.webp
var favicon []byte

var defaultWebServer = webSvr{
	engineFunc:   func() *gin.Engine { return gin.New() },
	readTimeout:  time.Second * 120,
	writeTimeout: time.Second * 120,
	idleTimeout:  time.Second * 60,
	hosts:        make([]string, 0),
	bind:         ":6880",
	protocol:     ProtocolHTTP,
	tlsc:         nil,
}

// Opt 通用化http框架
type webSvr struct {
	engineFunc   func() *gin.Engine
	hosts        []string
	tlsc         *tls.Config
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	bind         string
	protocol     ProtocolType
	enable       bool
}

func (opt *webSvr) build(l logger.Logger) (*http.Server, error) {
	if opt.bind == "" {
		return nil, errors.New("[web] services not enable")
	}
	s := &http.Server{
		Addr:         opt.bind,
		ReadTimeout:  opt.readTimeout,
		WriteTimeout: opt.writeTimeout,
		IdleTimeout:  opt.idleTimeout,
		TLSConfig:    opt.tlsc,
	}
	return s, nil
}

func (opt *webSvr) buildRoutes() (*gin.Engine, error) {
	h := opt.engineFunc()
	for _, v := range h.Routes() {
		if v.Path == "/favicon.ico" {
			return h, nil
		}
	}
	h.GET("/favicon.ico", func(c *gin.Context) {
		c.Writer.Write(favicon)
	})
	return h, nil
}

type webOpts func(opt *webSvr)

func OptWebEngineFunc(f func() *gin.Engine) webOpts {
	return func(opt *webSvr) {
		opt.engineFunc = f
	}
}

func OptWebHosts(hosts ...string) webOpts {
	return func(opt *webSvr) {
		opt.hosts = hosts
	}
}

func OptWebTimeout(read, write, idle time.Duration) webOpts {
	return func(opt *webSvr) {
		opt.readTimeout = read
		opt.writeTimeout = write
		opt.idleTimeout = idle
	}
}

func OptWebBind(s, cert, key string) webOpts {
	return func(opt *webSvr) {
		_, ok := checkTCPAddr(s)
		if !ok {
			opt.bind = ""
			return
		}
		if t, err := crypto.TLSConfigFromFile(cert, key, ""); err == nil {
			opt.bind = s
			opt.tlsc = t
			opt.protocol = ProtocolHTTPS
			return
		}
		opt.bind = s
		opt.protocol = ProtocolHTTP
	}
}
