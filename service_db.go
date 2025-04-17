package gofactory

import (
	"database/sql"

	"github.com/xyzj/toolbox/db"
	"gorm.io/gorm"
)

func (s *Service) DBQuery(sql string, rowcount int, args ...interface{}) (*db.QueryData, error) {
	return s.dbcli.Query(sql, rowcount, args...)
}

func (s *Service) DBExec(sql string, args ...interface{}) (int64, int64, error) {
	return s.dbcli.Exec(sql, args...)
}

func (s *Service) DBExecPrepare(sql string, args ...interface{}) error {
	return s.dbcli.ExecPrepare(sql, 0, args...)
}

func (s *Service) DBQueryBydb(dbidx int, sql string, rowcount int, args ...interface{}) (*db.QueryData, error) {
	return s.dbcli.QueryByDB(dbidx, sql, rowcount, args...)
}

func (s *Service) DBExecBydb(dbidx int, sql string, args ...interface{}) (int64, int64, error) {
	return s.dbcli.ExecByDB(dbidx, sql, args...)
}

func (s *Service) DBExecPrepareBydb(dbidx int, sql string, args ...interface{}) error {
	return s.dbcli.ExecPrepareByDB(dbidx, sql, 0, args...)
}

func (s *Service) DBOrm(dbidx int) (*gorm.DB, error) {
	dbidx = min(max(dbidx, 1), len(s.opt.clidb.database))
	return s.dbcli.ORM(dbidx)
}

// DBClient returns a *sql.DB connection for the specified database index.
// It ensures the database index is within the valid range before retrieving the SQL database client.
func (s *Service) DBClient(dbidx int) (*sql.DB, error) {
	dbidx = min(max(dbidx, 1), len(s.opt.clidb.database))
	return s.dbcli.SQLDB(dbidx)
}
