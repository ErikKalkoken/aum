package main

import (
	"encoding/json"
	"log"
	"net/http"

	"context"
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"

	"example/telemetry/queries"
)

//go:embed schema.sql
var ddl string

type Payload struct {
	UID  string
	Data string
}

func main() {
	ctx := context.Background()
	db, err := createDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	q := queries.New(db)

	http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		report := &Payload{}
		if err := json.NewDecoder(r.Body).Decode(report); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("got report:", report)

		err := q.CreateRecord(ctx, queries.CreateRecordParams{
			Uid:  report.UID,
			Data: report.Data,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("saved report:", report)

		w.WriteHeader(http.StatusCreated)
	})

	if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func createDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, err
	}
	return db, nil
}
