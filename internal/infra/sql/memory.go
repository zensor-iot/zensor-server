package sql

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

		startSIGHUPListener()
	})

	return &DB{DB: gormDB, autoMigrationEnabled: true, timeout: 0}, nil
}

func startSIGHUPListener() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP)

		for range sigChan {
			slog.Info("received SIGHUP signal, flushing in-memory database")
			if err := flushAllTables(); err != nil {
				slog.Error("failed to flush in-memory database", slog.Any("error", err))
			} else {
				slog.Info("successfully flushed in-memory database")
			}
		}
	}()
}

func flushAllTables() error {
	if gormDB == nil {
		slog.Warn("attempted to flush in-memory database but it was not initialized")
		return nil
	}

	// Get all table names
	var tables []string
	if err := gormDB.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables).Error; err != nil {
		return err
	}

	// Truncate all tables (SQLite uses DELETE FROM instead of TRUNCATE)
	for _, table := range tables {
		if err := gormDB.Exec("DELETE FROM " + table).Error; err != nil {
			slog.Error("failed to truncate table", slog.String("table", table), slog.Any("error", err))
			return err
		}
	}

	slog.Info("in-memory database flushed", slog.Int("tables_truncated", len(tables)))
	return nil
}
