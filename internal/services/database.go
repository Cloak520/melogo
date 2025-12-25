package services

import (
	"database/sql"
	"fmt"
	"melogo/internal/config"
	"melogo/internal/utils"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDatabase initializes the SQLite database
func InitDatabase(cfg *config.Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %v", err)
	}

	// Open database connection
	var err error
	DB, err = sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Configure connection pool
	if cfg.Database.MaxIdleConns > 0 {
		DB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	}
	if cfg.Database.MaxOpenConns > 0 {
		DB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime > 0 {
		DB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Minute)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	utils.NewLogger().Info("Database initialized successfully")
	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// 首先创建表
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			email VARCHAR(100) UNIQUE,
			avatar TEXT,
			avatar_blob BLOB,
			is_admin INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title VARCHAR(200) NOT NULL,
			artist VARCHAR(100),
			album VARCHAR(100),
			duration INTEGER,
			file_path TEXT NOT NULL,
			cover_image TEXT,
			lyrics_path TEXT,
			play_count INTEGER DEFAULT 0,
			is_collect INTEGER DEFAULT 0,
			is_deleted INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			user_id INTEGER REFERENCES users(id),
			is_public BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS playlist_songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER REFERENCES playlists(id) ON DELETE CASCADE,
			song_id INTEGER REFERENCES songs(id) ON DELETE CASCADE,
			order_index INTEGER,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			song_id INTEGER REFERENCES songs(id) ON DELETE CASCADE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS search_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			keyword VARCHAR(200) NOT NULL,
			searched_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS configurations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key VARCHAR(100) UNIQUE NOT NULL,
			value TEXT,
			description TEXT
		)`,
	}

	for _, tableSQL := range tables {
		_, err := DB.Exec(tableSQL)
		if err != nil {
			return fmt.Errorf("failed to create table or migrate: %v", err)
		}
	}

	return nil
}
