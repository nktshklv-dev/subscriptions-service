package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppAddr         string
	LogLevel        string
	ShutdownTimeout time.Duration

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
	DBSSL  string
}

func Load() (Config, error) {
	cfg := Config{
		AppAddr:         getEnv("APP_ADDR", ":8080"),
		LogLevel:        getEnv("APP_LOG_LEVEL", "debug"),
		ShutdownTimeout: getDuration("APP_SHUTDOWN_TIMEOUT", 5*time.Second),

		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASSWORD", "postgres"),
		DBName: getEnv("DB_NAME", "subscriptions"),
		DBSSL:  getEnv("DB_SSLMODE", "disable"),
	}

	return cfg, nil
}

func (c Config) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPass,
		c.DBName,
		c.DBSSL,
	)
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
	}
	return def
}
