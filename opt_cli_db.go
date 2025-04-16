package gofactory

import (
	"github.com/xyzj/toolbox/db"
)

type cliDB struct {
	cli  *db.Conn
	addr string
	user string
	pwd  string

	enable bool
}

type dbOpts func(o *cliDB)
