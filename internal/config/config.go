package config

import (
	"fmt"
	"os"
)

type Config struct {
	DSN        string  // postgres connection string
	ServerAddr string  // :8080
	FeeRate    float64 // commission rate
}

func Load() (*Config, error) {
	user, err := mustEnv("DB_USER")
	if err != nil {
		return nil, err
	}

	pass, err := mustEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	host, err := mustEnv("DB_HOST")
	if err != nil {
		return nil, err
	}

	port, err := mustEnv("DB_PORT")
	if err != nil {
		return nil, err
	}

	dbName, err := mustEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	sslMode := envOrDefault("DB_SSLMODE", "disable")
	serverAddr := envOrDefault("SERVER_ADDR", ":8080")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		pass,
		host,
		port,
		dbName,
		sslMode,
	)

	return &Config{
		DSN:        dsn,
		ServerAddr: serverAddr,
		FeeRate:    0.015,
	}, nil
}

func mustEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required env variable %q is not set", key)
	}

	return v, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
