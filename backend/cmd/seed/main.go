package main

import (
	"log"

	"boilerplate-learner/internal/config"
	"boilerplate-learner/internal/infrastructure/console/commands"
	"boilerplate-learner/internal/infrastructure/persistence"
	"boilerplate-learner/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	appLog, err := logger.New(cfg.Log.Level)
	if err != nil {
		log.Fatalf("logger error: %v", err)
	}
	defer func() { _ = appLog.Sync() }()

	db, err := persistence.NewDB(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	if err := commands.RunSeed(db, appLog, false); err != nil {
		log.Fatalf("seed error: %v", err)
	}
}
