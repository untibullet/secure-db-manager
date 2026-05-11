package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBName         string
	SuperuserUser  string
	SuperuserPass  string
	JWTSecret      string
	ServerAddr     string
	SessionTTL     time.Duration
	ReaperInterval time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		DBHost:         getenv("DB_HOST", "localhost"),
		DBPort:         getenv("DB_PORT", "5432"),
		DBName:         getenv("DB_NAME", "security_db"),
		SuperuserUser:  getenv("DB_SUPERUSER", "postgres"),
		SuperuserPass:  mustenv("DB_SUPERUSER_PASSWORD"),
		JWTSecret:      mustenv("JWT_SECRET"),
		ServerAddr:     getenv("SERVER_ADDR", ":8080"),
		SessionTTL:     parseDuration(getenv("SESSION_TTL", "8h")),
		ReaperInterval: parseDuration(getenv("REAPER_INTERVAL", "5m")),
	}

	if cfg.SuperuserPass == "" {
		return nil, fmt.Errorf("config: DB_SUPERUSER_PASSWORD is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}
	return cfg, nil
}

// SuperuserDSN возвращает строку подключения суперпользователя (AD-10).
// Не передавать в логи — содержит пароль.
func (c *Config) SuperuserDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.SuperuserUser, c.SuperuserPass, c.DBHost, c.DBPort, c.DBName)
}

// UserDSN возвращает строку подключения для конкретного пользователя (AD-2, AD-11).
// Не передавать в логи — содержит пароль.
func (c *Config) UserDSN(user, pass string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, c.DBHost, c.DBPort, c.DBName)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustenv(key string) string {
	return os.Getenv(key)
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 8 * time.Hour
	}
	return d
}
