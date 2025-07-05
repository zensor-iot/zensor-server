package sql

import (
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	databaseCreationOnce sync.Once
	gormDB               *gorm.DB
)

func NewMemoryORM(migrationsPath string) (ORM, error) {
	var err error
	databaseCreationOnce.Do(func() {
		dialector := sqlite.Open("file::memory:?cache=shared")
		gormDB, err = gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			panic(err)
		}
		db, err := gormDB.DB()
		if err != nil {
			panic(err)
		}
		db.SetMaxOpenConns(1)
	})

	return &DB{DB: gormDB, autoMigrationEnabled: true}, nil
}
