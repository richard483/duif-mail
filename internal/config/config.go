package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	// Global instance of config (singleton pattern)
	instance *Config
	once     sync.Once
)

// Config holds all application configuration
type Config struct {
	Server ServerConfig
	Mail   MailConfig
	App    AppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
}

// MailConfig holds mail service configuration
type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// AppConfig holds general application configuration
type AppConfig struct {
	Name     string
	Env      string
	LogLevel string
}

// Init initializes the global config instance
// Call this ONCE at application startup (in main.go)
func Init() error {
	var err error
	once.Do(func() {
		// Load .env file (optional - fails silently if not found)
		_ = godotenv.Load()

		instance = &Config{
			Server: ServerConfig{
				Port: getEnv("PORT", "50051"),
			},
			Mail: MailConfig{
				Host:     getEnv("MAIL_HOST", "smtp.gmail.com"),
				Port:     getEnvAsInt("MAIL_PORT", 587),
				Username: getEnv("MAIL_USERNAME", ""),
				Password: getEnv("MAIL_PASSWORD", ""),
			},
			App: AppConfig{
				Name:     getEnv("APP_NAME", "GoMail API"),
				Env:      getEnv("APP_ENV", "development"),
				LogLevel: getEnv("LOG_LEVEL", "info"),
			},
		}

		// Validate required fields
		err = instance.Validate()
	})
	return err
}

// Get returns the global config instance
// Make sure Init() is called first in main.go
func Get() *Config {
	if instance == nil {
		panic("config not initialized! Call config.Init() first in main.go")
	}
	return instance
}

// Load reads configuration from environment variables and returns it
// Use this if you prefer dependency injection instead of global access
func Load() (*Config, error) {
	// Load .env file (optional - fails silently if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "50051"),
		},
		Mail: MailConfig{
			Host:     getEnv("MAIL_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("MAIL_PORT", 587),
			Username: getEnv("MAIL_USERNAME", ""),
			Password: getEnv("MAIL_PASSWORD", ""),
		},
		App: AppConfig{
			Name:     getEnv("APP_NAME", "GoMail API"),
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.Mail.Username == "" {
		return fmt.Errorf("MAIL_USERNAME is required")
	}
	if c.Mail.Password == "" {
		return fmt.Errorf("MAIL_PASSWORD is required")
	}
	return nil
}

// IsDevelopment checks if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction checks if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// Helper functions

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer or returns default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool reads an environment variable as boolean or returns default
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
