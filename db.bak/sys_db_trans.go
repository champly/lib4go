package db

import "database/sql"

type ISysDBTrans interface {
	Query(query string, args ...interface{}) ([]QRow, error)
	ExecuteI(query string, args ...interface{}) (rowsaffected int64, lastinsertid int64, err error)
	Execute(query string, args ...interface{}) (rowsaffected int64, err error)
	Rollback() error
	Commit() error
}

type SysDBTrans struct {
	tx *sql.Tx
}

func (t *SysDBTrans) Query(query string, args ...interface{}) ([]QRow, error) {
	rows, err := t.tx.Query(query, args)
	if err != nil {
		return nil, err
	}
	return analyzeRows(rows)
}

func (t *SysDBTrans) ExecuteI(query string, args ...interface{}) (rowsaffected int64, lastinsertid int64, err error) {
	result, err := t.tx.Exec(query, args)
	if err != nil {
		return 0, 0, err
	}
	rowsaffected, err = result.RowsAffected()
	lastinsertid, err = result.LastInsertId()
	return
}

func (t *SysDBTrans) Execute(query string, args ...interface{}) (rowsaffected int64, err error) {
	result, err := t.tx.Exec(query, args)
	if err != nil {
		return 0, err
	}
	rowsaffected, err = result.RowsAffected()
	return
}

func (t *SysDBTrans) Rollback() error {
	return t.tx.Rollback()
}

func (t *SysDBTrans) Commit() error {
	return t.tx.Commit()
}
