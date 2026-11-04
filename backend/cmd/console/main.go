package main

import (
	"boilerplate-learner/internal/config"
	"boilerplate-learner/internal/infrastructure/console"
	"boilerplate-learner/internal/infrastructure/persistence"
	"boilerplate-learner/pkg/logger"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
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

	// La DB se abre solo cuando un comando la necesita, no para --help.
	db := lazyDB(cfg, log)

	kernel := console.NewKernel(cfg, log, db)
	return kernel.Execute()
}

// lazyDB devuelve una función que abre la conexión solo la primera vez que se llama.
func lazyDB(cfg *config.Config, log *zap.Logger) func() (*gorm.DB, error) {
	var (
		once sync.Once
		db   *gorm.DB
		err  error
	)
	return func() (*gorm.DB, error) {
		once.Do(func() {
			db, err = persistence.NewDB(cfg)
			if err != nil {
				log.Error("error conectando a la base de datos", zap.Error(err))
			}
		})
		return db, err
	}
}
