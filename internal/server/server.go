package server

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

	"example/telemetry/internal/model"
	"example/telemetry/internal/storage"
	"example/telemetry/internal/storage/queries"
)

var (
	//go:embed templates/*
	templatesFS embed.FS
	templates   map[string]*template.Template
)

// New create a new web server and returns it's main handler.
func New(db *sql.DB, q *queries.Queries) http.Handler {
	templates = make(map[string]*template.Template)
	router := http.NewServeMux()
	router.HandleFunc("/create-report", handleCreateRecord(q))
	// TODO: Add basic auth to status page
	router.HandleFunc("/show-status", handleShowStatus(q))
	return loggingMiddleware(router)
}

// LoadTemplates loads and parses all html templates.
func LoadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	files, err := fs.ReadDir(templatesFS, "templates")
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		t, err := template.ParseFS(templatesFS, "templates/"+f.Name(), "templates/base.html")
		if err != nil {
			return err
		}
		templates[f.Name()] = t
	}
	return nil
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
			t, ok := templates["status.html"]
			if !ok {
				return fmt.Errorf("status.html")
			}
			t.Execute(w, apps)
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
