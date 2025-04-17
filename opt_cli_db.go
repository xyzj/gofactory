package gofactory

import (
	"github.com/xyzj/toolbox/db"
	"github.com/xyzj/toolbox/logger"
)

type cliDB struct {
	cli      *db.Conn
	driver   db.Drive
	addr     string
	user     string
	pwd      string
	database []string
	enable   bool
}

func (opt *cliDB) build(l logger.Logger) error {
	if !opt.enable {
		return nil
	}
	var err error
	opt.cli, err = db.New(&db.Opt{
		Server:     opt.addr,
		User:       opt.user,
		Passwd:     opt.pwd,
		DriverType: opt.driver,
		Logger:     l,
	})
	return err
}

type dbOpts func(o *cliDB)

func OptDBHost(driver db.Drive, host, username, password string) dbOpts {
	return func(o *cliDB) {
		o.addr = host
		o.user = username
		o.pwd = password
		o.driver = driver
	}
}

func OptDBDatabases(databases ...string) dbOpts {
	return func(o *cliDB) {
		o.database = databases
	}
}
