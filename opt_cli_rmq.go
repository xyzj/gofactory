package gofactory

import (
	"crypto/tls"

	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/mq"
)

type cliRmq struct {
	clip            *mq.RMQProducer
	tlsc            *tls.Config
	recvFunc        func(topic string, body []byte)
	addr            string
	user            string
	pwd             string
	vhost           string
	exchange        string
	queueName       string
	queueDurable    bool // 队列是否持久化
	queueAutoDelete bool // 队列在不用时是否删除
	enableP         bool
	enableC         bool
	enable          bool
}

func (opt *cliRmq) build(l logger.Logger) error {
	if opt.enableP {
		opt.clip = mq.NewRMQProducer(&mq.RabbitMQOpt{
			Addr:         opt.addr,
			Username:     opt.user,
			Passwd:       opt.pwd,
			VHost:        opt.vhost,
			ExchangeName: opt.exchange,
			TLSConf:      opt.tlsc,
			LogHeader:    "[rmq-p]",
		}, l)
	}
	if opt.enableC {
		mq.NewRMQConsumer(&mq.RabbitMQOpt{
			Addr:            opt.addr,
			Username:        opt.user,
			Passwd:          opt.pwd,
			VHost:           opt.vhost,
			ExchangeName:    opt.exchange,
			QueueName:       opt.queueName,
			QueueDurable:    opt.queueDurable,
			QueueAutoDelete: opt.queueAutoDelete,
			TLSConf:         opt.tlsc,
			LogHeader:       "[rmq-c]",
		}, l, opt.recvFunc)
	}
	return nil
}

type rmqOpts func(o *cliRmq)

func OptRmqAuth(addr, host, user, pwd string, t *tls.Config) rmqOpts {
	return func(o *cliRmq) {
		if _, ok := checkTCPAddr(host); !ok {
			return
		}
		if t == nil {
			t = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		o.addr = addr
		o.user = user
		o.pwd = pwd
		o.vhost = host
		o.tlsc = t
	}
}

func OptRmqProducer(exchange string) rmqOpts {
	return func(o *cliRmq) {
		o.enableP = true
		o.exchange = exchange
	}
}

func OptRmqConsumer(exchange, queueName string, queueDurable, queueAutoDelete bool, recvFunc func(topic string, body []byte)) rmqOpts {
	return func(o *cliRmq) {
		o.enableC = true
		o.exchange = exchange
		o.recvFunc = recvFunc
		o.queueName = queueName
		o.queueDurable = queueDurable
		o.queueAutoDelete = queueAutoDelete
	}
}
