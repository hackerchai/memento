package main

import (
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/pkg/xlog"
)

// Embed the migrations directory (now local to this file's expected location)
//
//go:embed migrations
var migrationsFS embed.FS

// GetDatabaseDriverName maps config driver name to migrate driver name if necessary.
func GetDatabaseDriverName(driver string) string {
	switch driver {
	case "pg", "postgres", "postgresql":
		return "postgres"
	case "mysql", "mariadb":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite3"
	default:
		return driver
	}
}

// RunMigrations executes database migrations using embedded FS.
// Note: This function is now part of the main package.
func RunMigrations(cfg config.DatabaseConfig, logger *xlog.Logger) error {
	driverName := GetDatabaseDriverName(cfg.Driver)
	if driverName == "" {
		return fmt.Errorf("unsupported database driver for migration: %s", cfg.Driver)
	}

	migrationSubPath := "migrations/" + driverName // Path within the embedded migrationsFS

	// Create the iofs source instance using the subpath within the embedded FS
	sourceInstance, err := iofs.New(migrationsFS, migrationSubPath)
	if err != nil {
		logger.ErrorX(nil).Err(err).Str("path", migrationSubPath).Msg("Failed to create iofs source instance")
		return fmt.Errorf("failed to create iofs source instance: %w", err)
	}

	// Construct the database URL for migrate based on the driver
	var dbURL string
	switch driverName {
	case "postgres", "mysql":
		// Prepend scheme only if it's missing in the source DSN
		if !strings.HasPrefix(cfg.Source, driverName+"://") {
			dbURL = fmt.Sprintf("%s://%s", driverName, cfg.Source)
		} else {
			dbURL = cfg.Source // Assume source already has the correct scheme
		}
	case "sqlite3":
		// sqlite3 driver expects just the file path
		dbURL = cfg.Source
	default:
		return fmt.Errorf("cannot construct dbURL for unsupported driver: %s", driverName)
	}

	logger.InfoX(nil).Str("path", migrationSubPath).Str("driver", driverName).Str("dbURL", dbURL).Msg("Attempting to run database migrations from embedded FS") // Log the final URL

	// Create migrate instance using the source instance and the formatted dbURL
	m, err := migrate.NewWithSourceInstance("iofs", sourceInstance, dbURL)
	if err != nil {
		logger.ErrorX(nil).Err(err).Str("dbURL", dbURL).Msg("Failed to create migrate instance with iofs") // Log the URL used
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Apply migrations up
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.ErrorX(nil).Err(err).Msg("Failed to apply migrations")
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.InfoX(nil).Msg("No new migrations to apply")
	} else {
		logger.InfoX(nil).Msg("Database migrations applied successfully")
	}

	// Log current migration version
	version, dirty, err := m.Version()
	if err != nil {
		logger.ErrorX(nil).Err(err).Msg("Failed to get migration version")
	} else {
		logger.InfoX(nil).Uint("version", version).Bool("dirty", dirty).Msg("Current migration status")
		if dirty {
			logger.WarnX(nil).Msg("Migration state is dirty!")
		}
	}

	return nil
}
