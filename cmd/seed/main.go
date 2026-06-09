package main

import (
	"log"

	"cellar-backend/configs"
	"cellar-backend/internal/modules/user"
	"cellar-backend/pkg/auth"
	"cellar-backend/pkg/database"
	"cellar-backend/pkg/logger"

	gormlogger "gorm.io/gorm/logger"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.InitLogger(cfg.App.Env, cfg.App.LogLevel)

	db, err := database.InitDB(&cfg.DB, cfg.App.Env)
	if err != nil {
		logger.Fatal(err, "Failed to connect to database")
	}

	db.Logger = db.Logger.LogMode(gormlogger.Error)

	logger.Info("Starting database seeding...")

	hashedPassword, err := auth.HashPassword("password")
	if err != nil {
		logger.Fatal(err, "Failed to hash seed passwords")
	}

	seedUsers := []user.User{
		{
			FullName:   "System Admin",
			Email:      "admin@cellar.com",
			Password:   string(hashedPassword),
			Role:       "admin",
			IsVerified: true,
		},
		{
			FullName:   "Standard User A",
			Email:      "usera@cellar.com",
			Password:   string(hashedPassword),
			Role:       "user",
			IsVerified: true,
		},
		{
			FullName:   "Standard User B",
			Email:      "userb@cellar.com",
			Password:   string(hashedPassword),
			Role:       "user",
			IsVerified: true,
		},
		{
			FullName:   "Nadia Varra",
			Email:      "nadiavarra@yopmail.com",
			Password:   string(hashedPassword),
			Role:       "user",
			IsVerified: true,
		},
		{
			FullName:   "Rouf",
			Email:      "rouf@yopmail.com",
			Password:   string(hashedPassword),
			Role:       "user",
			IsVerified: true,
		},
	}

	for _, u := range seedUsers {
		var existing user.User
		err := db.First(&existing, "email = ?", u.Email).Error
		if err != nil {
			if err := db.Create(&u).Error; err != nil {
				logger.Error(err, "Failed to seed user: %s", u.Email)
			} else {
				logger.Info("Successfully seeded user: %s (%s)", u.Email, u.Role)
			}
		} else {
			logger.Info("User already exists, skipping: %s", u.Email)
		}
	}

	enforcer, err := auth.NewCasbinEnforcer("configs/rbac_model.conf", db)
	if err != nil {
		logger.Fatal(err, "Failed to initialize Casbin enforcer for seeding")
	}

	policies := [][]string{
		{"admin", "/api/v1/*", ".*"},
		{"user", "/api/v1/users/:id", "GET|PUT"},
		{"user", "/api/v1/health", "GET"},
	}

	for _, policy := range policies {
		added, err := enforcer.AddPolicy(policy[0], policy[1], policy[2])
		if err == nil && added {
			logger.Info("Successfully seeded Casbin policy: %v", policy)
		}
	}

	logger.Info("Database seeding completed successfully")
}
