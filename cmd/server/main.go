package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-crud-db-p2/config"
	"go-crud-db-p2/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig(".", "")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger, err := logger.New(logger.Config{
		Format:    cfg.Log.Format,
		FilePath:  cfg.Log.FilePath,
		Level:     cfg.Log.Level,
		ToStdout:  cfg.Log.ToStdout,
		AddSource: cfg.Log.AddSource,
	})
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}
	defer logger.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	defer cancel()

	handler, err := BootstrapApp(ctx, cfg)
	if err != nil {
		slog.Error("failed to bootstrap application", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:         cfg.Server.ListenAddress(),
		Handler:      handler.Engine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening", "address", srv.Addr, "driver", cfg.Database.Driver)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server crash", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("application has been shut down safely.")
}
