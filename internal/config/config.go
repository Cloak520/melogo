package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	AllowRegistration bool
	JWTSecret         string
}

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Music    MusicConfig
	Auth     AuthConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Host     string
	Port     int
	Debug    bool
	TLS      bool
	CertFile string
	KeyFile  string
}

// Address returns the server address in host:port format
func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Path            string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // in minutes
}

// MusicConfig holds the music configuration
type MusicConfig struct {
	Directory      string
	ScanInterval   int // in minutes
	AllowedFormats []string
	LyricsAPIURL   string
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() *Config {
	// Load default .env file if it exists
	godotenv.Load()

	return loadConfigFromEnv()
}

// LoadConfigFromFile loads configuration from a specified .env file
func LoadConfigFromFile(envPath string) *Config {
	// Load the specified .env file
	godotenv.Load(envPath)

	return loadConfigFromEnv()
}

// loadConfigFromEnv loads configuration from environment variables
func loadConfigFromEnv() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host:     getEnvOrDefault("SERVER_HOST", "localhost"),
			Port:     getEnvIntOrDefault("SERVER_PORT", 8080),
			Debug:    getEnvBoolOrDefault("SERVER_DEBUG", false),
			TLS:      getEnvBoolOrDefault("SERVER_TLS", false),
			CertFile: getEnvOrDefault("SERVER_CERT_FILE", ""),
			KeyFile:  getEnvOrDefault("SERVER_KEY_FILE", ""),
		},
		Database: DatabaseConfig{
			Path:            getEnvOrDefault("DATABASE_PATH", "./data/melogo.db"),
			MaxIdleConns:    getEnvIntOrDefault("DATABASE_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvIntOrDefault("DATABASE_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvIntOrDefault("DATABASE_CONN_MAX_LIFETIME", 60), // 60 minutes
		},
		Music: MusicConfig{
			Directory:      getEnvOrDefault("MUSIC_DIRECTORY", "./music"),
			ScanInterval:   getEnvIntOrDefault("MUSIC_SCAN_INTERVAL", 5), // 5 minutes
			AllowedFormats: []string{".mp3", ".wav", ".flac", ".m4a", ".aac", ".ogg"},
			LyricsAPIURL:   getEnvOrDefault("LYRICS_API_URL", "https://api.lrc.cx"),
		},
		Auth: AuthConfig{
			AllowRegistration: getEnvBoolOrDefault("ALLOW_REGISTRATION", true),
			JWTSecret:         getEnvOrDefault("JWT_SECRET", "melogo-secret-key-change-in-production"),
		},
	}

	// Ensure music directory exists
	if err := os.MkdirAll(cfg.Music.Directory, 0755); err != nil {
		fmt.Printf("Warning: Failed to create music directory: %v\n", err)
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create database directory: %v\n", err)
	}

	return cfg
}

// Helper functions to get environment variables with defaults
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
