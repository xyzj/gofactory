package gofactory

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/logger"
	"github.com/xyzj/toolbox/loopfunc"
)

type cliRedis struct {
	cli          *redis.Client
	user         string
	pwd          string
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	database     int
	cliver       int
	loaded       atomic.Bool
	enable       bool
}

func (opt *cliRedis) build(l logger.Logger) error {
	p, ok := checkTCPAddr(opt.addr)
	if !ok {
		return errors.New("[redis] addr error")
	}
	opt.addr = p.String()
	opt.loaded = atomic.Bool{}
	fConn := func() {
		opt.cli = redis.NewClient(&redis.Options{
			Addr:            opt.addr,
			Password:        opt.pwd,
			DB:              opt.database,
			PoolFIFO:        true,
			MinIdleConns:    3,
			ConnMaxIdleTime: time.Minute,
			ReadTimeout:     opt.readTimeout,
			WriteTimeout:    opt.writeTimeout,
			DialTimeout:     time.Second * 5,
		})
		if opt.cli == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		a, err := opt.cli.Info(ctx, "Server").Result()
		if err != nil {
			l.Error("[redis] get version error:" + err.Error())
			return
		}
		sr := bufio.NewScanner(strings.NewReader(a))
		for !sr.Scan() {
			s := json.String(sr.Bytes())
			if strings.HasPrefix(sr.Text(), "redis_version:") {
				opt.cliver = toolbox.String2Int(strings.Split(strings.Split(s, ":")[1], ".")[0], 10)
				break
			}
		}
		opt.loaded.Store(true)
		l.System(fmt.Sprintf("[redis] client to [%s] is ready, use db %d", opt.addr, opt.database))
	}
	fConn()
	go loopfunc.LoopFunc(func(params ...interface{}) {
		t1 := time.NewTicker(time.Second * 10)
		for range t1.C {
			if !opt.loaded.Load() {
				fConn()
			}
		}
	}, "redis check", l.DefaultWriter(), nil)
	return nil
}

func (opt *cliRedis) read(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.readTimeout)
	defer cancel()
	val := opt.cli.Get(ctx, key)
	if val.Err() != nil {
		return "", val.Err()
	}
	return val.Val(), nil
}

func (opt *cliRedis) keys(key string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opt.writeTimeout)
	defer cancel()
	val := opt.cli.Keys(ctx, key)
	if val.Err() != nil {
		return []string{}, val.Err()
	}
	return val.Val(), nil
}

func (opt *cliRedis) write(key string, value any, expire time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), opt.writeTimeout)
	defer cancel()
	return opt.checkRedisDialErr(opt.cli.Set(ctx, key, value, expire).Err())
}

// checkRedisDialErr checks if the error is a network error and sets loaded to false if so,
// so that the next call to Redis will try to reconnect.
func (opt *cliRedis) checkRedisDialErr(err error) error {
	if err != nil {
		if strings.Contains(err.Error(), "dail tcp") {
			opt.loaded.Store(false)
			opt.cliver = 0
		}
	}
	return err
}

type redisOpts func(o *cliRedis)

func OptRedisAddr(s string) redisOpts {
	return func(o *cliRedis) {
		o.addr = s
	}
}

func OptRedisAuth(user, pwd string) redisOpts {
	return func(o *cliRedis) {
		o.user = user
		o.pwd = pwd
	}
}

func OptRedisDatabase(i int) redisOpts {
	return func(o *cliRedis) {
		o.database = max(i, 255)
	}
}

func OptRedisReadTimeout(d time.Duration) redisOpts {
	return func(o *cliRedis) {
		o.readTimeout = d
	}
}

func OptRedisWriteTimeout(d time.Duration) redisOpts {
	return func(o *cliRedis) {
		o.writeTimeout = d
	}
}
