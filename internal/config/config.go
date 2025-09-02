package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppAddr         string
	LogLevel        string
	ShutdownTimeout time.Duration
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() (Config, error) {
	port, _ := strconv.Atoi(env("DB_PORT", "5432"))
	timeout, _ := time.ParseDuration(env("APP_SHUTDOWN_TIMEOUT", "5s"))

	cfg := Config{
		AppAddr:         env("APP_ADDR", ":8080"),
		LogLevel:        env("LOG_LEVEL", "info"),
		ShutdownTimeout: timeout,
		DBHost:          env("DB_HOST", "localhost"),
		DBPort:          port,
		DBUser:          env("DB_USER", ""),
		DBPassword:      env("DB_PASSWORD", ""),
		DBName:          env("DB_NAME", ""),
		DBSSLMode:       env("DB_SSL_MODE", "disable"),
	}

	missing := []string{}

	if cfg.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if cfg.DBPort == 0 {
		missing = append(missing, "DB_PORT")
	}
	if cfg.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if cfg.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if cfg.DBName == "" {
		missing = append(missing, "DB_NAME")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env variables:\n%s", strings.Join(missing, ",\n"))
	}

	return cfg, nil
}
