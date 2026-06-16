// Package db opens a GORM connection via a pluggable driver registry. SQLite
// (pure-Go, no CGO) and PostgreSQL ship by default; Register adds more.
package db

import (
	"fmt"
	"sort"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DialectorFactory builds a GORM dialector from a DSN.
type DialectorFactory func(dsn string) gorm.Dialector

// registry maps driver names to their dialector factories.
var registry = map[string]DialectorFactory{
	"sqlite":   func(dsn string) gorm.Dialector { return sqlite.Open(dsn) },
	"postgres": func(dsn string) gorm.Dialector { return postgres.Open(dsn) },
}

// Register adds or overrides a database driver. Call before Open.
func Register(name string, f DialectorFactory) { registry[name] = f }

// Drivers returns the sorted list of registered driver names.
func Drivers() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Open connects to the database using the named driver.
func Open(driver, dsn string) (*gorm.DB, error) {
	factory, ok := registry[driver]
	if !ok {
		return nil, fmt.Errorf("unknown db driver %q (registered: %v)", driver, Drivers())
	}

	gdb, err := gorm.Open(factory(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", driver, err)
	}

	// Enforce foreign keys on SQLite (off by default).
	if driver == "sqlite" {
		if err := gdb.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
		}
	}
	return gdb, nil
}

// Migrate runs GORM auto-migration for the supplied models.
func Migrate(gdb *gorm.DB, models ...any) error {
	if err := gdb.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}
