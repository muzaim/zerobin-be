package main

import (
	"log"

	"cellar-backend/configs"
	"cellar-backend/internal/modules/user"
	"cellar-backend/pkg/database"
	"cellar-backend/pkg/logger"

	"golang.org/x/crypto/bcrypt"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// Initialize configurations
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.InitLogger(cfg.App.Env, cfg.App.LogLevel)

	// Open DB connection
	db, err := database.InitDB(&cfg.DB, cfg.App.Env)
	if err != nil {
		logger.Fatal(err, "Failed to connect to database")
	}

	// Disable verbose SQL logging for seeding command
	db.Logger = db.Logger.LogMode(gormlogger.Error)

	logger.Info("Starting database seeding...")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		logger.Fatal(err, "Failed to hash seed passwords")
	}

	seedUsers := []user.User{
		{
			FullName: "System Admin",
			Email:    "admin@cellar.com",
			Password: string(hashedPassword),
			Role:     "admin",
		},
		{
			FullName: "Content Moderator",
			Email:    "moderator@cellar.com",
			Password: string(hashedPassword),
			Role:     "moderator",
		},
		{
			FullName: "Standard User A",
			Email:    "usera@cellar.com",
			Password: string(hashedPassword),
			Role:     "user",
		},
		{
			FullName: "Standard User B",
			Email:    "userb@cellar.com",
			Password: string(hashedPassword),
			Role:     "user",
		},
	}

	for _, u := range seedUsers {
		var existing user.User
		err := db.First(&existing, "email = ?", u.Email).Error
		if err != nil {
			// Record doesn't exist, create it
			if err := db.Create(&u).Error; err != nil {
				logger.Error(err, "Failed to seed user: %s", u.Email)
			} else {
				logger.Info("Successfully seeded user: %s (%s)", u.Email, u.Role)
			}
		} else {
			logger.Info("User already exists, skipping: %s", u.Email)
		}
	}

	logger.Info("Database seeding completed successfully")
}
