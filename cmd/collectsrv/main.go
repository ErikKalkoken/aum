// Collectsrv is a web service for collecting reports about application usage.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"example/telemetry/internal/server"
	"example/telemetry/internal/storage"
	"example/telemetry/internal/storage/queries"
)

var portFlag = flag.String("port", "8000", "port on which the server listens")
var createAppFlag = flag.Bool("create-app", false, "update or create app")

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
	db, err := storage.InitDB(ctx, "db.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	q := queries.New(db)

	if *createAppFlag {
		return createApp(ctx, q)
	}

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", port),
		Handler: server.New(db, q),
	}
	if err := server.LoadTemplates(); err != nil {
		log.Fatal(err)
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

func createApp(ctx context.Context, q *queries.Queries) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter the ID of the app: ")
	scanner.Scan()
	appID := scanner.Text()
	fmt.Print("Enter the name of the app: ")
	scanner.Scan()
	name := scanner.Text()
	found, err := storage.ApplicationExists(ctx, q, appID)
	if err != nil {
		return err
	}
	var msg string
	if found {
		msg = "Update existing app"
	} else {
		msg = "Create new app"
	}
	arg := queries.UpdateOrCreateApplicationParams{AppID: appID, Name: name}
	fmt.Printf("App: %+v - %s? (y/N)? ", arg, msg)
	var input string
	fmt.Scan(&input)
	if input != "y" {
		return nil
	}
	if err := q.UpdateOrCreateApplication(ctx, arg); err != nil {
		return err
	}
	fmt.Println("App created")
	return nil
}
