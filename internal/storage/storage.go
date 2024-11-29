package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"example/telemetry/internal/storage/queries"
	"fmt"
	"log/slog"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

func InitDB(ctx context.Context, dsn string) (*sql.DB, error) {
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

func ApplicationExists(ctx context.Context, q *queries.Queries, appID string) (bool, error) {
	n, err := q.CountApplicationByID(ctx, appID)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// TruncateTables will purge data from all tables. This is meant for tests.
func TruncateTables(db *sql.DB) {
	_, err := db.Exec("PRAGMA foreign_keys = 0")
	if err != nil {
		panic(err)
	}
	sql := `SELECT name FROM sqlite_master WHERE type = "table"`
	rows, err := db.Query(sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		tables = append(tables, name)
	}
	for _, n := range tables {
		sql := fmt.Sprintf("DELETE FROM %s;", n)
		_, err := db.Exec(sql)
		if err != nil {
			panic(err)
		}
	}
	// for _, n := range tables {
	// 	sql := fmt.Sprintf("DELETE FROM SQLITE_SEQUENCE WHERE name='%s'", n)
	// 	_, err := db.Exec(sql)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	_, err = db.Exec("PRAGMA foreign_keys = 1")
	if err != nil {
		panic(err)
	}
}
