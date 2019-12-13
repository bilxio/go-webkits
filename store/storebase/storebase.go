package storebase

import (
	"context"
	"database/sql"

	"github.com/bilxio/go-webkits/store/db"
)

// ToParamsFunc convert model to params
type ToParamsFunc func(obj interface{}) map[string]interface{}

// ScanRowFunc scan a sql row into desternation model object
type ScanRowFunc func(scanner db.Scanner, dst interface{}) error

// ScanRowsFunc scans rows
type ScanRowsFunc func(rows *sql.Rows) ([]interface{}, error)

// StoreBase implements bascis CRUD methods
type StoreBase struct {
	db       *db.DB
	toParams ToParamsFunc
	scanRow  ScanRowFunc
	scanRows ScanRowsFunc
}

// New create a storeBase with db instance and other helper funcs
func New(
	db *db.DB,
	toParams ToParamsFunc,
	scanRow ScanRowFunc,
	scanRows ScanRowsFunc) *StoreBase {

	return &StoreBase{db, toParams, scanRow, scanRows}
}

// GenericFind ...
func (s *StoreBase) GenericFind(ctx context.Context, sql string, out interface{}) error {
	return s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		params := s.toParams(out)
		query, args, err := binder.BindNamed(sql, params)
		if err != nil {
			return err
		}
		row := queryer.QueryRow(query, args...)
		return s.scanRow(row, out)
	})
}

// GenericCreate ...
func (s *StoreBase) GenericCreate(ctx context.Context, sql string, obj interface{}) (int64, error) {
	var id int64
	err := s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params := s.toParams(obj)
		stmt, args, err := binder.BindNamed(sql, params)
		if err != nil {
			return err
		}
		res, err := execer.Exec(stmt, args...)
		if err != nil {
			return err
		}
		id, err = res.LastInsertId()
		return err
	})
	return id, err
}

// GenericList ...
func (s *StoreBase) GenericList(ctx context.Context, sql string, paramObject interface{}) ([]interface{}, error) {
	var out []interface{}
	err := s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		params := s.toParams(paramObject)
		stmt, args, err := binder.BindNamed(sql, params)
		if err != nil {
			return err
		}
		rows, err := queryer.Query(stmt, args...)
		if err != nil {
			return err
		}
		out, err = s.scanRows(rows)
		return err
	})
	return out, err
}

// GenericDelete ...
func (s *StoreBase) GenericDelete(ctx context.Context, sql string, model interface{}) error {
	return s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params := s.toParams(model)
		stmt, args, err := binder.BindNamed(sql, params)
		if err != nil {
			return err
		}
		_, err = execer.Exec(stmt, args...)
		return err
	})
}

// GenericUpdate ...
func (s *StoreBase) GenericUpdate(ctx context.Context, sql string, model interface{}) error {
	return s.db.Lock(func(execer db.Execer, binder db.Binder) error {
		params := s.toParams(model)
		stmt, args, err := binder.BindNamed(sql, params)
		if err != nil {
			return err
		}
		_, err = execer.Exec(stmt, args...)
		return err
	})
}

// GenericCount ...
func (s *StoreBase) GenericCount(ctx context.Context, sql string) (int64, error) {
	var out int64
	err := s.db.View(func(queryer db.Queryer, binder db.Binder) error {
		return queryer.QueryRow(sql).Scan(&out)
	})
	return out, err
}
