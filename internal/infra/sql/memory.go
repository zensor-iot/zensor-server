package sql

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewMemoryORM(migrationsPath string, replacements map[string]string) (ORM, error) {
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite in-memory db: %w", err)
	}

	return &DB{DB: gormDB, autoMigrationEnabled: true}, nil
}
