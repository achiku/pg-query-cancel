package main

// https://github.com/golang/go/wiki/SQLInterface

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
)

func NewDB(url string) (DBer, error) {
	cfg, err := pgx.ParseConnectionString(url)
	if err != nil {
		return nil, err
	}
	dbCfg := &stdlib.DriverConfig{
		ConnConfig: cfg,
	}
	stdlib.RegisterDriverConfig(dbCfg)
	db, err := sql.Open("pgx", dbCfg.ConnectionString(""))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Second * 60)
	return db, nil
}

// https://pkg.go.dev/database/sql
// Queryer database/sql compatible query interface
type Queryer interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Txer database/sql transaction interface
type Txer interface {
	Queryer
	Commit() error
	Rollback() error
}

// DBer database/sql
type DBer interface {
	Queryer
	Begin() (*sql.Tx, error)
	Close() error
	Ping() error
}

func main() {
	fmt.Println("query cancel test")
}
