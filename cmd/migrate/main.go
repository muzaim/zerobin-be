package main

import (
	"log"
	"os"

	"cellar-backend/configs"
	"cellar-backend/internal/modules/audit"
	"cellar-backend/internal/modules/user"
	"cellar-backend/pkg/database"
	"cellar-backend/pkg/logger"
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

	args := os.Args
	if len(args) < 2 {
		log.Fatal("Migration command required: up, rollback, or automigrate")
	}

	command := args[1]
	switch command {
	case "up", "automigrate":
		logger.Info("Starting database migration (Up)...")
		err := db.AutoMigrate(
			&audit.AuditTrail{},
			&user.User{},
		)
		if err != nil {
			logger.Fatal(err, "Database migration failed")
		}
		logger.Info("Database migration completed successfully")

	case "rollback":
		logger.Info("Starting database rollback (Down/Reset)...")
		err := db.Migrator().DropTable(
			&user.User{},
			&audit.AuditTrail{},
		)
		if err != nil {
			logger.Fatal(err, "Database rollback failed")
		}
		logger.Info("Database rollback completed successfully")

	default:
		log.Fatalf("Unknown migration command: %s. Use up, rollback, or automigrate", command)
	}
}
