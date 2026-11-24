package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("file .env tidak ditemukan, memakai environment sistem")
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	if n, err := strconv.Atoi(GetEnv(key, "")); err == nil && n > 0 {
		return n
	}
	return fallback
}
