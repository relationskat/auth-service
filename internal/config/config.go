package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DB       *DB
	Cache    *Cache
	Provider *Provider
	SMTP     *SMTP
	GRPC     *GRPC
}

func New() (*Config, error) {
	cfg := &Config{
		DB: &DB{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     uint(getEnvInt("DB_PORT", 5432)),
			User:     getEnv("DB_USER", "auth"),
			Password: getEnv("DB_PASSWORD", "auth_password"),
			DB:       getEnv("DB_NAME", "auth"),
		},
		Cache: &Cache{
			Host: getEnv("CACHE_HOST", "localhost"),
			Port: getEnvInt("CACHE_PORT", 11211),
		},
		Provider: &Provider{
			PublicKeyPath:   getEnv("PROVIDER_PUBLIC_KEY_PATH", "infra/keys/public.pem"),
			PrivateKeyPath:  getEnv("PROVIDER_PRIVATE_KEY_PATH", "infra/keys/private.pem"),
			AccessTokenTTL:  getEnvInt("PROVIDER_ACCESS_TOKEN_TTL", 86400),
			RefreshTokenTTL: getEnvInt("PROVIDER_REFRESH_TOKEN_TTL", 2592000),
		},
		SMTP: &SMTP{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnvInt("SMTP_PORT", 1025),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "no-reply@auth.local"),
		},
		GRPC: &GRPC{
			Host:            getEnv("GRPC_HOST", "0.0.0.0"),
			Port:            getEnv("GRPC_PORT", "50051"),
			RequestTimeout:  time.Duration(getEnvInt("GRPC_REQUEST_TIMEOUT_SECONDS", 15)) * time.Second,
			ResponseTimeout: time.Duration(getEnvInt("GRPC_RESPONSE_TIMEOUT_SECONDS", 15)) * time.Second,
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
