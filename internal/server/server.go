package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"example/telemetry/internal/model"
	"example/telemetry/internal/storage"
	"example/telemetry/internal/storage/queries"
)

func New(db *sql.DB, q *queries.Queries) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/create-report", handleCreateRecord(q))
	// TODO: Add basic auth to status page
	router.HandleFunc("/show-status", handleShowStatus(q))
	return loggingMiddleware(router)
}

// TODO: Generate HTML pages from templates

func handleShowStatus(q *queries.Queries) http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		err := func() error {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return nil
			}
			apps, err := q.ListApplications(ctx)
			if err != nil {
				return err
			}
			html := "<table><tr><th>ID</th><th>Name</th></tr>\n"
			for _, a := range apps {
				html += fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", a.AppID, a.Name)
			}
			html += "</table>\n"
			_, err = w.Write([]byte(html))
			return err
		}()
		if err != nil {
			slog.Error("show status", "error", err)
			http.Error(w, "failed to show status page", http.StatusInternalServerError)
		}
	}
}

func handleCreateRecord(q *queries.Queries) http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		err := func() error {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return nil
			}

			report := &model.ReportPayload{}
			if err := json.NewDecoder(r.Body).Decode(report); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return nil
			}

			found, err := storage.ApplicationExists(ctx, q, report.AppID)
			if err != nil {
				return err
			}
			if !found {
				http.Error(w, "bad application ID", http.StatusBadRequest)
				return nil
			}

			err = q.CreateReport(ctx, queries.CreateReportParams{
				AppID:     report.AppID,
				Arch:      report.Arch,
				MachineID: report.MachineID,
				Os:        report.OS,
				Version:   report.Version,
			})
			if err != nil {
				return err
			}
			slog.Info("report created", "report", report)

			w.WriteHeader(http.StatusCreated)
			return nil
		}()
		if err != nil {
			slog.Error("create report", "error", err)
			http.Error(w, "failed to create report", http.StatusInternalServerError)
		}
	}
}
