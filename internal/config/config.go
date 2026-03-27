package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	SuperAdminUsername string
	SuperAdminPassword string
	Env                string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/attendancemgmt?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production"),
		Port:               getEnv("PORT", "8080"),
		SuperAdminUsername: getEnv("SUPER_ADMIN_USERNAME", "sainigp20"),
		SuperAdminPassword: getEnv("SUPER_ADMIN_PASSWORD", ""),
		Env:                getEnv("ENV", "development"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
