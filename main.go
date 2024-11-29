package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"context"

	"example/telemetry/internal/model"
	"example/telemetry/queries"
)

var portFlag = flag.Int("port", 8000, "port on which the server listens")

func main() {
	flag.Parse()
	ctx := context.Background()
	db, err := createDB(ctx, "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	q := queries.New(db)

	router := http.NewServeMux()
	router.HandleFunc("/create-report", func(w http.ResponseWriter, r *http.Request) {
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
	})

	loggingMiddleware := newLoggingMiddleware(ctx)
	serverAddress := fmt.Sprintf("localhost:%d", *portFlag)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: loggingMiddleware(router),
	}
	go func() {
		slog.Info("Web server started", "address", serverAddress)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("Web server terminated prematurely", "error", err)
		}
	}()

	// Ensure graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Web server shutdown", "error", err)
	} else {
		slog.Info("Web server stopped")
	}
}
