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

	"example/telemetry/queries"
)

type Payload struct {
	UID  string
	Data string
}

var portFlag = flag.Int("port", 8000, "port on which the server listens")
var usernameFlag = flag.String("username", "app", "username for authenticating incoming requests")
var passwordFlag = flag.String("password", "", "password for authenticating incoming requests")

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
	router.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
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
	routerWithMiddleware := loggingMiddleware(ctx, authMiddleware(*usernameFlag, *passwordFlag, router))

	serverAddress := fmt.Sprintf("localhost:%d", *portFlag)
	server := &http.Server{
		Addr:    serverAddress,
		Handler: routerWithMiddleware,
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
