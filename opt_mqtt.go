package gofactory

import (
	"crypto/tls"
	"errors"
	"log/slog"
	"time"

	"github.com/xyzj/mqtt-server/cmd/server"
	"github.com/xyzj/mqtt-server/hooks/auth"
	"github.com/xyzj/toolbox/crypto"
	"github.com/xyzj/toolbox/logger"
)

type mqttBroker struct {
	tlsc              *tls.Config
	auth              *auth.Ledger
	maxMsgExpiry      time.Duration // max message expiry time in seconds
	maxSessionExpiry  time.Duration // max session expiry time in seconds
	mqtt              string
	mqtttls           string
	mqttweb           string
	mqttws            string
	clientsBufferSize int // clients read and write buffer size in bytes
	insidejob         bool
	enable            bool
}

func (opt *mqttBroker) build(l logger.Logger, mode RunMode) (*server.MqttServer, error) {
	if !opt.enable {
		return nil, errors.New("[mqtt] broker not enable")
	}
	ll := slog.LevelInfo
	as := true
	if mode == Release {
		ll = slog.LevelWarn
		as = false
	}
	mopt := &server.Opt{
		ClientsBufferSize:       opt.clientsBufferSize,
		MaxMsgExpirySeconds:     int(opt.maxMsgExpiry.Seconds()),
		MaxSessionExpirySeconds: int(opt.maxSessionExpiry.Seconds()),
		InsideJob:               opt.insidejob,
		AuthConfig:              opt.auth,
		DisableAuth:             opt.auth == nil,
		FileLogger: slog.New(slog.NewTextHandler(
			l.DefaultWriter(),
			&slog.HandlerOptions{
				AddSource: as,
				Level:     ll,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == "time" {
						return slog.Attr{}
					}
					return a
				},
			},
		)),
	}
	if opt.tlsc == nil || opt.tlsc.Certificates == nil {
		opt.mqtttls = ""
	}
	_, ok1 := checkTCPAddr(opt.mqtt)
	_, ok2 := checkTCPAddr(opt.mqtttls)
	if !ok1 && !ok2 {
		return nil, errors.New("[mqtt] broker error: no valid ports")
	}
	return server.NewServer(mopt), nil
}

var defaultMqttBroker = mqttBroker{
	mqtt:              ":1883",
	mqtttls:           "",
	mqttweb:           ":1880",
	mqttws:            "",
	tlsc:              nil,
	maxMsgExpiry:      time.Hour,
	maxSessionExpiry:  time.Second * 360,
	clientsBufferSize: 4096,
	auth:              nil,
	insidejob:         false,
}

type mqttBrokerOpts func(o *mqttBroker)

func OptMqttAddr(s string) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.mqtt = s
	}
}

func OptMqttTlsAddr(s string, t *tls.Config) mqttBrokerOpts {
	return func(o *mqttBroker) {
		if t == nil || t.Certificates == nil {
			o.mqtttls = ""
			return
		}
		o.mqtttls = s
		o.tlsc = t
	}
}

func OptMqttTlsFromFile(s, cert, key, ca string) mqttBrokerOpts {
	return func(o *mqttBroker) {
		t, err := crypto.TLSConfigFromFile(cert, key, ca)
		if err != nil {
			o.mqtttls = ""
		} else {
			OptMqttTlsAddr(s, t)
		}
	}
}

func OptMqttWebAddr(s string) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.mqttweb = s
	}
}

func OptMqttWSAddr(s string) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.mqttws = s
	}
}

func OptMqttMaxMsgExpirySeconds(t time.Duration) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.maxMsgExpiry = min(t, time.Hour*24*7)
	}
}

func OptMqttMaxSessionExpirySeconds(t time.Duration) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.maxSessionExpiry = min(t, time.Hour*24*7)
	}
}

func OptMqttClientBufferSize(s int) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.clientsBufferSize = min(max(s, 1024), 16384)
	}
}

func OptMqttInsideClient(b bool) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.insidejob = true
	}
}

func OptMqttAuthConfig(ac *auth.Ledger) mqttBrokerOpts {
	return func(o *mqttBroker) {
		o.auth = ac
	}
}

func OptMqttAuthFromfile(f string) mqttBrokerOpts {
	return func(o *mqttBroker) {
		if ac, err := server.FromAuthfile(f, false); err != nil {
			o.auth = nil
		} else {
			o.auth = ac
		}
	}
}
