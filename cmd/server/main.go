package main

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/noahxsu777/Gooo/internal/app"
	"github.com/noahxsu777/Gooo/internal/config"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	service := app.NewService(cfg, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", service.Healthz)
	mux.HandleFunc("/readyz", service.Readyz)
	mux.HandleFunc("/metrics", service.Metrics)
	mux.HandleFunc("/api/state", service.HandleState)
	mux.HandleFunc("/api/control", service.HandleControl)
	mux.HandleFunc("/webhook/tiktools", service.HandleWebhook)
	mux.HandleFunc("/events", service.ServeEvents)
	mux.Handle("/", http.FileServer(http.Dir("web/static")))

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      logMiddleware(logger, corsMiddleware(cfg.AllowedOrigin, mux)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	service.Start(ctx)

	go func() {
		logger.Info("server starting", "addr", cfg.Addr, "mode", cfg.Mode, "channel", cfg.Channel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	service.Stop()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("server stopped")
}

func corsMiddleware(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigin == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tiktools-Signature")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
