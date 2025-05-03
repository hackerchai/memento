package db

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"

	_ "github.com/go-sql-driver/mysql" // Import mysql driver
)

// NewDBConnection creates a new database connection based on the driver and source.
func NewDBConnection(driver, source string, verbose bool) (*bun.DB, error) {
	var sqldb *sql.DB
	var err error

	switch driver {
	case "pg", "postgres", "postgresql":
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(source)))
	case "mysql", "mariadb":
		sqldb, err = sql.Open("mysql", source)
		if err != nil {
			return nil, fmt.Errorf("failed to open mysql connection: %w", err)
		}
	case "sqlite", "sqlite3":
		sqldb, err = sql.Open(sqliteshim.ShimName, source)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}
		// SQLite specific settings for better performance/concurrency
		sqldb.SetMaxOpenConns(1) // Important for SQLite to avoid database locking issues
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Check the connection
	if err := sqldb.Ping(); err != nil {
		sqldb.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	var db *bun.DB
	switch driver {
	case "pg", "postgres", "postgresql":
		db = bun.NewDB(sqldb, pgdialect.New())
	case "mysql", "mariadb":
		db = bun.NewDB(sqldb, mysqldialect.New())
	case "sqlite", "sqlite3":
		db = bun.NewDB(sqldb, sqlitedialect.New())
	}

	// Add a query hook for logging SQL queries if verbose is true.
	if verbose {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return db, nil
}
