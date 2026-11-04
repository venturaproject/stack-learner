package persistence

import (
	"boilerplate-learner/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Silent
	if cfg.Server.Mode == "debug" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	if err := Migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}
