package main

import (
	"context"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	url := "postgres://chiku:@localhost:5432/postgres"
	db, err := NewDB(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	t.Log("ok")
}

func TestQueryContext(t *testing.T) {
	url := "postgres://chiku:@localhost:5432/postgres"
	db, err := NewDB(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	var tm time.Time
	if err := db.QueryRowContext(ctx, "select now()").Scan(&tm); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", tm)
}

func TestQueryContextCancel(t *testing.T) {
	url := "postgres://chiku:@localhost:5432/postgres"
	db, err := NewDB(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var v interface{}
	if err := db.QueryRowContext(ctx, "select pg_sleep(10)").Scan(&v); err != nil {
		t.Fatalf("pg_sleep query failed: %s", err)
	}
	t.Logf("%s", v)
}
