package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPass       string
	DBName       string
	RedisAddr    string
	SupplierAURL string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			log.Println("no .env file, reading from environment")
		} else {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	dbPass, err := getRequiredEnv("DB_PASS")
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		Port:         getEnv("APP_PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPass:       dbPass,
		DBName:       getEnv("DB_NAME", "supplier_bridge"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		SupplierAURL: getEnv("SUPPLIER_A_URL", "http://localhost:9090"),
	}

	if err := validatePort("APP_PORT", cfg.Port); err != nil {
		return nil, err
	}
	if err := validatePort("DB_PORT", cfg.DBPort); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func validatePort(name, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s must be a number: %w", name, err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be in range 1-65535", name)
	}
	return nil
}

func getRequiredEnv(key string) (string, error) {
	if val := os.Getenv(key); val != "" {
		return val, nil
	}
	return "", fmt.Errorf("environment variable %s is required", key)
}
