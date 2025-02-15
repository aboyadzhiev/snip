package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/aboyadzhiev/snip/server/internal/handler"
	"github.com/aboyadzhiev/snip/server/internal/service"
	"github.com/aboyadzhiev/snip/server/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
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
	getenv func(string) string,
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

	db, err := initDB(ctx, getenv)
	if err != nil {
		return err
	}
	defer db.Close()

	valkeyClient, err := initValkeyClient(ctx, getenv)
	if err != nil {
		return err
	}
	defer valkeyClient.Close()

	logger := slog.New(slog.NewJSONHandler(stdout, nil))

	validate := initValidator()

	hostname := strings.TrimSpace(getenv("SNIP_HOSTNAME"))
	shortener, err := initURLShortener(logger, hostname, valkeyClient, db)
	if err != nil {
		return err
	}

	srv := NewServer(logger, validate, shortener)

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

func NewServer(logger *slog.Logger, validate *validator.Validate, shortener service.URLShortener) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByRealIP(30, 1*time.Minute))
	// Limit the max request body size to 1MB
	r.Use(middleware.RequestSize(1_048_576))

	addRoutes(r, logger, validate, shortener)

	var httpHandler http.Handler = r

	return httpHandler
}

func addRoutes(r *chi.Mux, _ *slog.Logger, validate *validator.Validate, shortener service.URLShortener) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", handler.Healthz())
		r.Post("/shortened-url", handler.ShortenURL(shortener, validate))
	})

	r.Route("/{slug}", func(r chi.Router) {
		r.Use(middleware.NoCache)
		r.Get("/", handler.Resolve(shortener))
	})

	r.Handle("/", http.NotFoundHandler())
}

func initValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return validate
}

func initURLShortener(_ *slog.Logger, hostname string, valkeyClient valkey.Client, db *pgxpool.Pool) (service.URLShortener, error) {
	sequence := store.NewShortenedURLSequence(valkeyClient)
	shortenedURLStore := store.NewShortenedURL(db)
	shortener := service.NewURLShortener(hostname, sequence, shortenedURLStore)

	return shortener, nil
}

func initDB(ctx context.Context, getenv func(string) string) (*pgxpool.Pool, error) {
	user := strings.TrimSpace(getenv("POSTGRES_USER"))
	password, err := readSecretFile(getenv("POSTGRES_PASSWORD_FILE"))
	if err != nil {
		return nil, err
	}
	hostname := strings.TrimSpace(getenv("POSTGRES_HOST"))
	database := strings.TrimSpace(getenv("POSTGRES_DB"))

	url := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, hostname, database)
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initValkeyClient(_ context.Context, getenv func(string) string) (valkey.Client, error) {
	valkeyHosts := strings.Split(getenv("VALKEY_HOSTS"), ",")
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{InitAddress: valkeyHosts})
	if err != nil {
		return nil, err
	}

	return valkeyClient, nil
}

func readSecretFile(file string) (string, error) {
	secret, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(secret)), nil
}
