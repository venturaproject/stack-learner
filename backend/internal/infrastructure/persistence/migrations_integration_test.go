//go:build integration

package persistence

import "testing"

func TestMigrationsAreRepeatable(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	var count int64
	if err := db.Table("schema_migrations").Count(&count).Error; err != nil {
		t.Fatalf("read migration history: %v", err)
	}
	if count < 3 {
		t.Fatalf("migration history has %d entries, want baseline and at least two SQL migrations", count)
	}
}
