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
	http:         ":6880",
	protocols:    map[string]string{"http": "http://127.0.0.1:6880"},
	tlsc:         nil,
}

// Opt 通用化http框架
type webSvr struct {
	engineFunc   func() *gin.Engine
	hosts        []string
	protocols    map[string]string
	tlsc         *tls.Config
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	http         string
	https        string
	enable       bool
}

func (opt *webSvr) build(l logger.Logger) (map[string]*http.Server, error) {
	if opt.http+opt.https == "" {
		return nil, errors.New("[web] services not enable")
	}
	m := make(map[string]*http.Server)
	if opt.https != "" {
		s := &http.Server{
			Addr:         opt.https,
			ReadTimeout:  opt.readTimeout,
			WriteTimeout: opt.writeTimeout,
			IdleTimeout:  opt.idleTimeout,
			TLSConfig:    opt.tlsc,
		}
		m["https"] = s
	}
	if opt.http != "" {
		s := &http.Server{
			Addr:         opt.http,
			ReadTimeout:  opt.readTimeout,
			WriteTimeout: opt.writeTimeout,
			IdleTimeout:  opt.idleTimeout,
		}
		m["http"] = s
	}
	return m, nil
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

func OptHTTP(s string) webOpts {
	return func(opt *webSvr) {
		if _, ok := checkTCPAddr(s); !ok {
			opt.http = ""
			delete(opt.protocols, "http")
			return
		}
		opt.http = s
		opt.protocols["http"] = "http://" + s
	}
}

func OptHTTPS(s string, t *tls.Config) webOpts {
	return func(opt *webSvr) {
		if _, ok := checkTCPAddr(s); !ok {
			opt.https = ""
			return
		}
		if t == nil || t.Certificates == nil {
			opt.https = ""
			return
		}
		opt.https = s
		opt.protocols["https"] = "https://" + s
		opt.tlsc = t
	}
}

func OptHTTPSFromFile(s, certFile, keyFile string) webOpts {
	return func(opt *webSvr) {
		t, err := crypto.TLSConfigFromFile(certFile, keyFile, "")
		if err != nil {
			opt.tlsc = nil
			return
		}
		OptHTTPS(s, t)
	}
}
