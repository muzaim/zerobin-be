package configs

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	JWT   JWTConfig
	CORS  CORSConfig
	MinIO MinIOConfig
	SMTP  SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Sender   string
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	PublicURL       string
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

	viper.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	viper.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	viper.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	viper.SetDefault("MINIO_USE_SSL", "false")
	viper.SetDefault("MINIO_BUCKET_NAME", "cellar-avatars")
	viper.SetDefault("MINIO_PUBLIC_URL", "http://localhost:9000/cellar-avatars")

	viper.SetDefault("SMTP_HOST", "localhost")
	viper.SetDefault("SMTP_PORT", 1025)
	viper.SetDefault("SMTP_USER", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_SENDER", "noreply@cellar.com")

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
		MinIO: MinIOConfig{
			Endpoint:        viper.GetString("MINIO_ENDPOINT"),
			AccessKeyID:     viper.GetString("MINIO_ACCESS_KEY"),
			SecretAccessKey: viper.GetString("MINIO_SECRET_KEY"),
			UseSSL:          viper.GetBool("MINIO_USE_SSL"),
			BucketName:      viper.GetString("MINIO_BUCKET_NAME"),
			PublicURL:       viper.GetString("MINIO_PUBLIC_URL"),
		},
		SMTP: SMTPConfig{
			Host:     viper.GetString("SMTP_HOST"),
			Port:     viper.GetInt("SMTP_PORT"),
			User:     viper.GetString("SMTP_USER"),
			Password: viper.GetString("SMTP_PASSWORD"),
			Sender:   viper.GetString("SMTP_SENDER"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.App.Env != "development" && c.App.Env != "production" && c.App.Env != "test" {
		return fmt.Errorf("invalid APP_ENV '%s': must be one of 'development', 'production', 'test'", c.App.Env)
	}

	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET must not be empty")
	}
	if c.App.Env == "production" && c.JWT.Secret == "super_secret_key_change_me_in_production_12345" {
		return errors.New("JWT_SECRET must be changed from the default value in production environment")
	}

	if c.DB.Host == "" {
		return errors.New("DB_HOST must not be empty")
	}
	if c.DB.Port == "" {
		return errors.New("DB_PORT must not be empty")
	}
	if c.DB.User == "" {
		return errors.New("DB_USER must not be empty")
	}
	if c.DB.Name == "" {
		return errors.New("DB_NAME must not be empty")
	}

	if c.MinIO.Endpoint == "" {
		return errors.New("MINIO_ENDPOINT must not be empty")
	}
	if c.MinIO.AccessKeyID == "" {
		return errors.New("MINIO_ACCESS_KEY must not be empty")
	}
	if c.MinIO.SecretAccessKey == "" {
		return errors.New("MINIO_SECRET_KEY must not be empty")
	}
	if c.MinIO.BucketName == "" {
		return errors.New("MINIO_BUCKET_NAME must not be empty")
	}

	if c.SMTP.Host == "" {
		return errors.New("SMTP_HOST must not be empty")
	}
	if c.SMTP.Port <= 0 {
		return fmt.Errorf("invalid SMTP_PORT '%d': must be greater than 0", c.SMTP.Port)
	}

	return nil
}
