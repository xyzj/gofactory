package main

import (
	"github.com/xyzj/gofactory"
	"github.com/xyzj/toolbox/db"
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
		gofactory.WithMQTTBroker(
			gofactory.OptMqttAddr(":6828"),
			gofactory.OptMqttWebAddr(":6829"),
		),
		gofactory.WithTCPServer(
			gofactory.OptTCPBindAddr(":6823"),
		),
		gofactory.WithRedisClient(
			gofactory.OptRedisAddr("192.168.50.83:6379"),
			gofactory.OptRedisAuth("", "arbalest"),
		),
		gofactory.WithWebServer(
			gofactory.OptWebBind(":6824", "", "")),
		gofactory.WithDiscover(gofactory.OptDiscoverByRedis(
			gofactory.OptRedisAddr("192.168.50.83:6379"),
			gofactory.OptRedisAuth("", "arbalest")),
			gofactory.OptDiscoverInfo("testsss", "测试", "127.0.0.1", "/wlst-micro", map[gofactory.ProtocolType]string{gofactory.ProtocolHTTP: "http://1222"}),
		),
		gofactory.WithBoltDB("test.db"),
		gofactory.WithMqttClient(gofactory.OptMqttAuth("arx7", "arbalest"),
			gofactory.OptMqttHost("tls://192.168.50.83:1881", nil)),
		gofactory.WithDBClient(gofactory.OptDBHost(db.DriveMySQL, "192.168.50.83:13306", "root", "lp1234xy")),
	)
	if err != nil {
		panic(err)
	}
	s.Run()
}
