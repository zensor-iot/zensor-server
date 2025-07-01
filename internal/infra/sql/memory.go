package sql

import (
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	databaseCreationOnce sync.Once
	gormDB               *gorm.DB
)

func NewMemoryORM(migrationsPath string, replacements map[string]string) (ORM, error) {
	var err error
	databaseCreationOnce.Do(func() {
		dialector := sqlite.Open("file::memory:?cache=shared")
		gormDB, err = gorm.Open(dialector, &gorm.Config{})
		db, err := gormDB.DB()
		if err != nil {
			panic(err)
		}
		db.SetMaxOpenConns(1)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite in-memory db: %w", err)
	}
	return &DB{DB: gormDB, autoMigrationEnabled: true}, nil
}
