package db

import (
	"time"
)

type DB struct {
	db ISysDB
}

func NewDB(dbtype, connectStr string, maxIdle, maxOpen, maxLifetime int) (obj *DB, err error) {
	obj = &DB{}
	obj.db, err = NewSysDB(dbtype, connectStr, maxIdle, maxOpen, time.Duration(maxLifetime)*time.Second)
	return
}

func (d *DB) Query(sql string, input map[string]interface{}) (data []QRow, query string, args []interface{}, err error) {
	// query, args = d
	return
}
