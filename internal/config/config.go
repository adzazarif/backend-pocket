package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	AppEnv         string
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	JWTSecret      string
	JWTExpiryHours int
	SeedUserEmail  string
	SeedUserPass   string
	SeedUserName   string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	return &Config{
		AppPort:        getEnv("APP_PORT", "3000"),
		AppEnv:         getEnv("APP_ENV", "development"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBName:         getEnv("DB_NAME", "pocket_app"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", "secret"),
		JWTSecret:      getEnv("JWT_SECRET", "secret"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		SeedUserEmail:  getEnv("SEED_USER_EMAIL", "user@example.com"),
		SeedUserPass:   getEnv("SEED_USER_PASSWORD", "password123"),
		SeedUserName:   getEnv("SEED_USER_NAME", "John Doe"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
