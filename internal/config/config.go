package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	SupabaseURL        string
	SupabaseServiceKey string
	PaystackSecretKey  string
	GinMode            string
	Port               string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (good for local development)
	err := godotenv.Load()
	if err != nil {
		// This is not necessarily fatal if env vars are set externally (e.g., in Docker, hosting platform)
		log.Printf("INFO: No .env file found or error loading .env: %v. Relying on system environment variables.", err)
	} else {
		log.Println("--- Successfully called godotenv.Load() ---")
	}

	cfg := &Config{
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_KEY"),
		PaystackSecretKey:  os.Getenv("PAYSTACK_SECRET_KEY"),
		GinMode:            os.Getenv("GIN_MODE"),
		Port:               os.Getenv("PORT"),
	}

	if cfg.SupabaseURL == "" {
		log.Fatal("SUPABASE_URL environment variable is required")
	}
	if cfg.SupabaseServiceKey == "" {
		log.Fatal("SUPABASE_SERVICE_KEY environment variable is required")
	}
	if cfg.PaystackSecretKey == "" {
		log.Fatal("PAYSTACK_SECRET_KEY environment variable is required")
	}

	if cfg.GinMode == "" {
		cfg.GinMode = "debug" // default to debug mode
	}
	if cfg.Port == "" {
		cfg.Port = "8080" // default port
	}
	// Validate port is a number
	if _, err := strconv.Atoi(cfg.Port); err != nil {
		log.Fatalf("Invalid PORT: %s. Must be a number.", cfg.Port)
	}


	return cfg, nil
}

const (
	// DATABYTES_PER_DATACREDIT_KOBO defines how many Databytes a user gets for 1 unit of Datacredit (which is 1 kobo).
	// Example: If 1 kobo buys 100 Databytes.
	DATABYTES_PER_DATACREDIT_KOBO = 100 // Adjust this value as per your business model
)