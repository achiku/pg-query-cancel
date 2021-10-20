package main

// https://github.com/golang/go/wiki/SQLInterface

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
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

// App application
type App struct {
	DB DBer
}

func newTimeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}

func (app *App) hello(w http.ResponseWriter, r *http.Request) {
	log.Println("hello access")
	fmt.Fprintf(w, "hello\n")
	return
}

func (app *App) slowHello(w http.ResponseWriter, r *http.Request) {
	log.Println("slow hello access")
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("waiting %d...", i)
	}
	fmt.Fprintf(w, "slow hello\n")
	return
}

func (app *App) slowHelloDB(w http.ResponseWriter, r *http.Request) {
	log.Println("slow hello db access")
	var val interface{}
	err := app.DB.QueryRowContext(r.Context(), "select pg_sleep(5)").Scan(&val)
	if err != nil {
		log.Println("query failed")
		return
	}
	fmt.Fprintf(w, "slow hello db\n")
	return
}

var (
	uri = flag.String("dburi", "", "database uri")
)

func main() {
	flag.Parse()

	db, err := NewDB(*uri)
	if err != nil {
		log.Fatal(err)
	}
	app := App{DB: db}

	timeout := newTimeoutMiddleware(1 * time.Second)
	mux := http.NewServeMux()
	mux.Handle("/hello", http.HandlerFunc(app.hello))
	mux.Handle("/slow-hello", http.HandlerFunc(app.slowHello))
	mux.Handle("/timeout-hello-custom", timeout(http.HandlerFunc(app.slowHello)))
	mux.Handle("/timeout-hello-stdlib", http.TimeoutHandler(http.HandlerFunc(app.slowHello), 2*time.Second, "timeout!!\n"))
	mux.Handle("/timeout-hello-db", http.TimeoutHandler(http.HandlerFunc(app.slowHelloDB), 2*time.Second, "timeout!!\n"))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
