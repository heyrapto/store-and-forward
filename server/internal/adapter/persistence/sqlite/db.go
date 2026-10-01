package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "github.com/mattn/go-sqlite3" // For Episode 1, we use mattn/go-sqlite3. 
)

// Open opens a SQLite database connection with WAL mode enabled.
// It also automatically runs goose migrations from the provided path.
func Open(dbPath, migrationsPath string) (*sql.DB, error) {
	// Enable WAL mode and foreign keys in the connection string
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000", dbPath)
	
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}

	// Set connection pool limits appropriate for SQLite WAL
	db.SetMaxOpenConns(1) // Mattn driver requires max 1 for writes to avoid locked DB issues easily, or we handle busy errors.
	
	// Run migrations
	goose.SetDialect("sqlite3")
	if err := goose.Up(db, migrationsPath); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}
