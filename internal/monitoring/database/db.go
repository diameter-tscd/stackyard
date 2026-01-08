package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB initializes the SQLite database for user settings
func InitDB() error {
	dbPath := "monitoring_users.db"

	// Ensure database file exists
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS user_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL DEFAULT 'Admin',
		photo_path TEXT,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// EnsureUploadDirectory creates the upload directory if it doesn't exist
func EnsureUploadDirectory(uploadDir string) error {
	profilesDir := filepath.Join(uploadDir, "profiles")
	return os.MkdirAll(profilesDir, 0755)
}
