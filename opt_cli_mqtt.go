package gofactory

import (
	"crypto/tls"
	"time"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/mq"
)

type cliMqtt struct {
	cli                    *mq.MqttClientV5
	tlsc                   *tls.Config                     // tls配置，默认为 InsecureSkipVerify: true
	subscribe              map[string]byte                 // 订阅消息，map[topic]qos
	sendTimeo              time.Duration                   // 发送超时
	clientID               string                          // ClientID 客户端标示，会添加随机字符串尾巴，最大22个字符
	addr                   string                          // 服务端ip:port
	user                   string                          // 登录用户名
	pwd                    string                          // 登录密码
	failureCacheMax        int                             // 最大缓存消息数量，默认10000
	failureCacheExpire     time.Duration                   // 消息缓存时间，默认一小时
	failureCacheExpireFunc func(topic string, body []byte) // 消息失效的处置方法
	recvFunc               func(topic string, body []byte) // 消息接收处置方法
	enableFailureCache     bool                            // 是否启用断连消息暂存
	enable                 bool
}

func (opt *cliMqtt) build(l logger.Logger) error {
	var err error
	opt.cli, err = mq.NewMQTTClientV5(&mq.MqttOpt{
		Logg:                   l,
		Username:               opt.user,
		Passwd:                 opt.pwd,
		ClientID:               opt.clientID,
		Addr:                   opt.addr,
		TLSConf:                opt.tlsc,
		SendTimeo:              opt.sendTimeo,
		Subscribe:              opt.subscribe,
		EnableFailureCache:     opt.enableFailureCache,
		FailureCacheMax:        opt.failureCacheMax,
		FailureCacheExpire:     opt.failureCacheExpire,
		FailureCacheExpireFunc: opt.failureCacheExpireFunc,
		LogHeader:              "[mqtt]",
	}, opt.recvFunc)
	if err != nil {
		return err
	}
	return nil
}

type mqttOpts func(o *cliMqtt)

func OptMqttHost(s string, t *tls.Config) mqttOpts {
	return func(o *cliMqtt) {
		o.addr = s
		if t == nil {
			t = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		o.tlsc = t
	}
}

func OptMqttClientID(s string) mqttOpts {
	return func(o *cliMqtt) {
		o.clientID = s
	}
}

func OptMqttAuth(user, pwd string) mqttOpts {
	return func(o *cliMqtt) {
		o.user = user
		o.pwd = pwd
	}
}

func OptMqttSendTimeout(d time.Duration) mqttOpts {
	return func(o *cliMqtt) {
		o.sendTimeo = d
	}
}

func OptMqttSubscribe(s map[string]byte) mqttOpts {
	return func(o *cliMqtt) {
		o.subscribe = s
	}
}

func OptMqttFailureCache(max int, expire time.Duration, expireFunc func(topic string, body []byte)) mqttOpts {
	return func(o *cliMqtt) {
		o.enableFailureCache = true
		o.failureCacheMax = max
		o.failureCacheExpire = expire
		o.failureCacheExpireFunc = expireFunc
	}
}

func OptMqttRecvFunc(f func(topic string, body []byte)) mqttOpts {
	return func(o *cliMqtt) {
		o.recvFunc = f
	}
}
