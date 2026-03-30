package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigins string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Env:         getEnv("ENV", "development"),
		Port:        getEnv("PORT", "3080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
