// @title Learner Learner API
// @version 1.0
// @description Backend Go/Gin con Clean Architecture + DDD
// @host localhost
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "boilerplate-learner/docs"
	"boilerplate-learner/internal/config"
	"boilerplate-learner/internal/di"
	"boilerplate-learner/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log, err := logger.New(cfg.Log.Level)
	if err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	container, err := di.Build(cfg, log)
	if err != nil {
		return fmt.Errorf("container: %w", err)
	}
	defer func() {
		if container.Redis != nil {
			if err := container.Redis.Close(); err != nil {
				log.Error("redis close error", zap.Error(err))
			}
		}
	}()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      container.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}
	log.Info("server stopped")
	return nil
}
