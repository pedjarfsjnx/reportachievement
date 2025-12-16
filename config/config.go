package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	PostgresDSN string
	MongoURI    string
	MongoDBName string
	JWTSecret   string
}

func LoadConfig() *Config {
	// Load .env
	_ = godotenv.Load()

	return &Config{
		AppPort: getEnv("APP_PORT", ":3000"),
		// Default DSN disesuaikan dengan setting lokal umumnya
		PostgresDSN: getEnv("DB_DSN", "host=localhost user=postgres password=pedja12345 dbname=report_achievement_db port=5432 sslmode=disable"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "achievement_logs"),
		JWTSecret:   getEnv("JWT_SECRET", "rahasia_negara"), // Default key jika env kosong
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
