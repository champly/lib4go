package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ISysDB interface {
	Query(query string, args ...interface{}) ([]QRow, error)
	ExecuteI(query string, args ...interface{}) (rowsaffected int64, lastinsertid int64, err error)
	Execute(query string, args ...interface{}) (rowsaffected int64, err error)
	Begin() (ISysDBTrans, error)
	Close() error
}

type SysDB struct {
	dbtype     string
	connectStr string
	db         *sql.DB
}

func NewSysDB(dbtype, connectStr string, maxIdle, maxOpen int, maxLifetime time.Duration) (*SysDB, error) {
	db, err := sql.Open(dbtype, connectStr)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)
	db.SetConnMaxLifetime(maxLifetime)
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &SysDB{dbtype, connectStr, db}, nil
}

func (d *SysDB) Query(query string, args ...interface{}) ([]QRow, error) {
	rows, err := d.db.Query(query, args)
	if err != nil {
		return nil, err
	}
	return analyzeRows(rows)
}

func (d *SysDB) ExecuteI(query string, args ...interface{}) (rowsaffected int64, lastinsertid int64, err error) {
	result, err := d.db.Exec(query, args)
	if err != nil {
		return 0, 0, err
	}
	rowsaffected, err = result.RowsAffected()
	lastinsertid, err = result.LastInsertId()
	return
}

func (d *SysDB) Execute(query string, args ...interface{}) (rowsaffected int64, err error) {
	result, err := d.db.Exec(query, args)
	if err != nil {
		return 0, err
	}
	rowsaffected, err = result.RowsAffected()
	return
}

func (d *SysDB) Begin() (ISysDBTrans, error) {
	t := &SysDBTrans{}
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	t.tx = tx
	return t, nil
}

func (d *SysDB) Close() error {
	return d.db.Close()
}

type QRow map[string]string

func analyzeRows(rows *sql.Rows) ([]QRow, error) {
	data := []QRow{}
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read query row fail:%s", err)
	}

	l := len(cols)
	for i := 0; i < l; i++ {
		cols[i] = strings.ToLower(cols[i])
	}

	values := make([][]byte, l)
	scans := make([]interface{}, l)
	for i := range values {
		scans[i] = &values[i]
	}
	for rows.Next() {
		row := map[string]string{}
		err := rows.Scan(scans...)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		for k, v := range values {
			row[cols[k]] = string(v)
		}
		data = append(data, row)
	}
	return data, nil
}
