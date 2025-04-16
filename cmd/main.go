package main

import (
	"time"

	"github.com/xyzj/gofactory"
	"github.com/xyzj/toolbox/gocmd"
	"github.com/xyzj/toolbox/logger"
)

func main() {
	gocmd.DefaultProgram(&gocmd.Info{
		Title: "gofactory",
		Ver:   "v0.1.0",
	}).Execute()
	s, err := gofactory.New(
		gofactory.WithLogger(logger.NewConsoleLogger()),
		gofactory.SetMode(gofactory.Release),
		gofactory.WithTCPServer(
			gofactory.OptTCPBindAddr(":6823"),
		),
		gofactory.WithRedisClient(
			gofactory.OptRedisAddr("192.168.50.83:6379"),
			gofactory.OptRedisAuth("", "arbalest"),
		),
		gofactory.WithWebServer(
			gofactory.OptHTTP(":6824")),
		gofactory.WithDiscover(gofactory.OptDiscoverByRedis(
			gofactory.OptRedisAddr("192.168.50.83:6379"),
			gofactory.OptRedisAuth("", "arbalest")),
			gofactory.OptDiscoverInfo("testsss", "测试", "127.0.0.1", "/wlst-micro", map[gofactory.ProtocolType]string{gofactory.ProtocolHTTP: "http://1222"}),
		),
		gofactory.WithBoltDB("test.db"),
		gofactory.WithMqttClient(gofactory.OptMqttAuth("arx7", "arbalest"),
			gofactory.OptMqttHost("tls://192.168.50.83:1881", nil)),
	)
	if err != nil {
		panic(err)
	}
	go func() {
		time.Sleep(time.Second * 5)
	}()
	s.Run()
}
