package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

func createDB(ctx context.Context, dsn string) (*sql.DB, error) {
	v := url.Values{}
	v.Add("_fk", "on")
	v.Add("_journal_mode", "WAL")
	v.Add("_busy_timeout", "5000") // 5000 = 5 seconds
	v.Add("_cache_size", "-20000") // -20000 = 20 MB
	v.Add("_synchronous", "normal")
	v.Add("_txlock", "IMMEDIATE")
	dsn2 := fmt.Sprintf("%s?%s", dsn, v.Encode())
	slog.Debug("Connecting to sqlite", "dsn", dsn2)
	db, err := sql.Open("sqlite3", dsn2)
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, err
	}
	return db, nil
}
