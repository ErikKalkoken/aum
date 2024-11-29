package main

import (
	"example/telemetry/queries"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"context"
)

var portFlag = flag.String("port", "8000", "port on which the server listens")

func main() {
	flag.Parse()
	ctx := context.Background()
	if err := run(ctx, *portFlag); err != nil {
		slog.Error("run failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, port string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	db, err := createDB(ctx, "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", port),
		Handler: newServer(queries.New(db)),
	}
	go func() {
		slog.Info("server started", "listening", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listening and serving", "error", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutting down server", "error", err)
		} else {
			slog.Info("server shut down complete")
		}
	}()
	wg.Wait()
	return nil
}
