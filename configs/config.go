package configs

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App  AppConfig
	DB   DBConfig
	JWT  JWTConfig
	CORS CORSConfig
}

type AppConfig struct {
	Env      string
	Port     string
	Name     string
	LogLevel string
}

type DBConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
	SSLMode     string
	AutoMigrate bool
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists (useful for local development)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file found but failed to load:", err)
		}
	}

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_NAME", "cellar-backend")
	viper.SetDefault("LOG_LEVEL", "info")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "cellar")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("AUTO_MIGRATE", "true")

	viper.SetDefault("JWT_SECRET", "super_secret_key_change_me_in_production_12345")
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")

	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")

	// Read environment variables
	viper.AutomaticEnv()

	// Parse duration variables safely
	accessExpiryDuration, err := time.ParseDuration(viper.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		log.Printf("Warning: Failed to parse JWT_ACCESS_EXPIRY, defaulting to 15 minutes: %v", err)
		accessExpiryDuration = 15 * time.Minute
	}

	refreshExpiryDuration, err := time.ParseDuration(viper.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		log.Printf("Warning: Failed to parse JWT_REFRESH_EXPIRY, defaulting to 7 days: %v", err)
		refreshExpiryDuration = 7 * 24 * time.Hour
	}

	// Parse CORS allowed origins
	allowedOriginsRaw := viper.GetString("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsRaw == "*" {
		allowedOrigins = []string{"*"}
	} else if allowedOriginsRaw != "" {
		allowedOrigins = []string{allowedOriginsRaw} // For simplicity, handle standard comma-separated split or direct value
	} else {
		allowedOrigins = []string{"*"}
	}

	cfg := &Config{
		App: AppConfig{
			Env:      viper.GetString("APP_ENV"),
			Port:     viper.GetString("APP_PORT"),
			Name:     viper.GetString("APP_NAME"),
			LogLevel: viper.GetString("LOG_LEVEL"),
		},
		DB: DBConfig{
			Host:        viper.GetString("DB_HOST"),
			Port:        viper.GetString("DB_PORT"),
			User:        viper.GetString("DB_USER"),
			Password:    viper.GetString("DB_PASSWORD"),
			Name:        viper.GetString("DB_NAME"),
			SSLMode:     viper.GetString("DB_SSLMODE"),
			AutoMigrate: viper.GetBool("AUTO_MIGRATE"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  accessExpiryDuration,
			RefreshExpiry: refreshExpiryDuration,
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
	}

	return cfg, nil
}
