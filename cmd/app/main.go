package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"subscriptions-service/internal/config"
	"subscriptions-service/internal/data"
	"subscriptions-service/internal/subscriptions"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = os.Stderr.WriteString("config load error: " + err.Error() + "\n")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     parseLevel(cfg.LogLevel),
		AddSource: true,
	}))
	slog.SetDefault(logger)

	db, err := data.NewDB(cfg.ConnString(), 10, 5, 30*time.Minute)
	if err != nil {
		logger.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	{
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			logger.Error("db ping failed", "err", err)
			os.Exit(1)
		}
		logger.Info("db connected")
	}

	repo := subscriptions.NewRepository(db)
	subHandler := subscriptions.NewHandler(repo, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/health/db", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "db: down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("db: ok"))
	})
	mux.Handle("/openapi.yaml", http.FileServer(http.Dir(".")))
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("./swagger-ui"))))
	subHandler.RegisterMux(mux)

	srv := &http.Server{
		Addr:              cfg.AppAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting server", "addr", cfg.AppAddr, "log_level", cfg.LogLevel)
		logger.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	} else {
		logger.Info("server stopped")
	}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
