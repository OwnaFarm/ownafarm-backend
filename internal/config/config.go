package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App    AppConfig
	DB     DBConfig
	Valkey ValkeyConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type ValkeyConfig struct {
	Addr     string
	Password string
	DB       int
	TLS      bool
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}

func LoadConfig() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env not loaded (will use OS environment variables):", err)
	}

	valkeyDB, err := strconv.Atoi(getEnv("VALKEY_DB", "0"))
	if err != nil {
		log.Fatal("env: VALKEY_DB must be an integer")
	}

	valkeyTLS, err := strconv.ParseBool(getEnv("VALKEY_TLS", "false"))
	if err != nil {
		log.Fatal("env: VALKEY_TLS must be a boolean")
	}

	return &Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "postgres"),
		},
		Valkey: ValkeyConfig{
			Addr:     getEnv("VALKEY_ADDR", "localhost:6379"),
			Password: getEnv("VALKEY_PASSWORD", ""),
			DB:       valkeyDB,
			TLS:      valkeyTLS,
		},
	}
}
