package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/aboyadzhiev/snip/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	_ func(string) string,
	_ io.Reader,
	stdout, stderr io.Writer,
) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS]\n", args[0])
		fmt.Println("OPTIONS:")
		flags.PrintDefaults()
	}

	var (
		addr = flags.String("addr", ":8081", "A TCP address to listen on e.g. 127.0.0.1:8081")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(stdout, nil))

	srv := NewServer(logger)

	httpServer := &http.Server{
		Addr:         *addr,
		Handler:      srv,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(fmt.Sprintf("Listening and serving on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_, _ = fmt.Fprintf(stderr, "Error listening and serving: %s\n", err)
		}
		logger.Info("Stopped serving new connections")
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, shutdownCtxRelease := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCtxRelease()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			_, _ = fmt.Fprintf(stderr, "Error shutting down: %s\n", err)
			if err = httpServer.Close(); err != nil {
				_, _ = fmt.Fprintf(stderr, "Error closing: %s\n", err)
			}
		}
		logger.Info("Shutdown completed")
	}()

	wg.Wait()

	return nil
}

func NewServer(logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByRealIP(30, 1*time.Minute))
	// Limit the max request body size to 1MB
	r.Use(middleware.RequestSize(1_048_576))

	// TODO: Use the `middleware.NoCache` when redirecting the user to the original URL
	// r.Use(middleware.NoCache)

	addRoutes(r, logger)

	var httpHandler http.Handler = r

	return httpHandler
}

func addRoutes(r *chi.Mux, _ *slog.Logger) {
	r.Get("GET /api/v1/healthz", handler.Healthz())

	r.Handle("/", http.NotFoundHandler())
}
