package main

import (
	"context"
	"encoding/json"
	"example/telemetry/internal/model"
	"example/telemetry/queries"
	"log/slog"
	"net/http"
)

func newServer(q *queries.Queries) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/create-report", handleCreateRecord(q))
	return loggingMiddleware(router)
}

func handleCreateRecord(q *queries.Queries) http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		report := &model.ReportPayload{}
		if err := json.NewDecoder(r.Body).Decode(report); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := q.CreateReport(ctx, queries.CreateReportParams{
			AppID:     report.AppID,
			Arch:      report.Arch,
			MachineID: report.MachineID,
			Os:        report.OS,
			Version:   report.Version,
		})
		if err != nil {
			slog.Error("store report", "error", err)
			http.Error(w, "failed to store report", http.StatusInternalServerError)
			return
		}
		slog.Info("report created", "report", report)

		w.WriteHeader(http.StatusCreated)
	}
}
