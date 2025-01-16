package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"passvault/config"
	"passvault/internal/clients/sso/grpc"
	"passvault/internal/http-server/handlers/client/register"
	"passvault/internal/http-server/handlers/entry/get"
	"passvault/internal/http-server/handlers/entry/save"
	authrest "passvault/internal/http-server/middlewares/auth"
	mwLogger "passvault/internal/http-server/middlewares/logger"
	"passvault/internal/lib/logger/sl"
	storage "passvault/internal/storage/sqlite"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

const (
	grpcHost = "localhost"
)

func main() {
	cfg := config.MustLoad()

	db, err := storage.New(cfg.StoragePath)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	log := setupLogger(cfg.Env)

	timeout := cfg.HTTPServer.Timeout * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	idleTimeout := cfg.HTTPServer.IdleTimeout * time.Second
	if idleTimeout == 0 {
		idleTimeout = 60 * time.Second
	}

	log = log.With(slog.String("env", cfg.Env))

	log.Info("initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")

	ctx, grpcClient, err := grpc.New(log, fmt.Sprintf("localhost:%d", cfg.GRPC.Port), cfg.GRPC.Timeout, cfg.GRPC.RetriesCount)
	if err != nil {
		log.Error("failed to create gRPC client", "error", err)
		os.Exit(1)
	}

	router := chi.NewRouter()

	authMiddleware := authrest.New(slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	), cfg.Secret)

	router.Use(authMiddleware)
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/save", save.New(log, db, timeout))

	router.Get("/get/{entry_id}", get.New(log, db, timeout))

	router.Get("/list", get.New(log, db, timeout))

	router.Get("/register", register.New(log, grpcClient, timeout))

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  idleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	if err := db.Close(); err != nil {
		log.Error("failed to stop storage", sl.Err(err))
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
