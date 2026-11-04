package persistence

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type legacyOrganization struct {
	ID        int32     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"not null;size:160"`
	Slug      string    `gorm:"not null;size:100;uniqueIndex"`
	CreatedBy int32     `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (legacyOrganization) TableName() string { return "organizations" }

// Migrate bootstraps a schema created before migration tracking existed, then
// applies each embedded, numbered SQL migration exactly once.
func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return migrate(tx)
	})
}

func migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version bigint PRIMARY KEY,
		name text NOT NULL,
		applied_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	if err := db.Exec(`SELECT pg_advisory_xact_lock(hashtext('boilerplate-learner-schema-migrations'))`).Error; err != nil {
		return fmt.Errorf("lock schema migrations: %w", err)
	}

	var baseline int64
	if err := db.Table("schema_migrations").Where("version = 1").Count(&baseline).Error; err != nil {
		return err
	}
	if baseline == 0 {
		if err := migrateBaseline(db); err != nil {
			return fmt.Errorf("apply schema baseline: %w", err)
		}
		if err := db.Exec("INSERT INTO schema_migrations (version, name) VALUES (1, ?)", "legacy-schema-baseline").Error; err != nil {
			return err
		}
	}

	entries, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(entries)
	for _, path := range entries {
		name := strings.TrimPrefix(path, "migrations/")
		version, err := strconv.ParseInt(strings.SplitN(name, "_", 2)[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid migration filename %q: %w", name, err)
		}
		var applied int64
		if err := db.Table("schema_migrations").Where("version = ?", version).Count(&applied).Error; err != nil {
			return err
		}
		if applied > 0 {
			continue
		}
		sql, err := migrationFiles.ReadFile(path)
		if err != nil {
			return err
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if err := db.Exec("INSERT INTO schema_migrations (version, name) VALUES (?, ?)", version, name).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateBaseline(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{}, &models.UserSettings{}, &legacyOrganization{},
		&models.OrganizationMembership{}, &models.Course{}, &models.CourseModule{}, &models.Lesson{},
		&models.Role{}, &models.Permission{}, &models.RefreshToken{},
		&models.Trabajador{}, &models.EstadoTelefono{}, &models.Tipologia{}, &models.Telefono{},
		&models.ActuacionTelefono{}, &models.DeviceBrand{}, &models.DeviceModel{}, &models.Dispositivo{},
		&models.ApiClient{}, &models.ApiClientToken{}, &models.RevokedAccessToken{}, &models.OutboxMessage{},
	)
}
