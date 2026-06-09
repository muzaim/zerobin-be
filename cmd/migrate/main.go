package main

import (
	"log"
	"os"

	"cellar-backend/configs"
	"cellar-backend/internal/modules/audit"
	"cellar-backend/internal/modules/blog"
	"cellar-backend/internal/modules/user"
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

	args := os.Args
	if len(args) < 2 {
		log.Fatal("Migration command required: up, rollback, or automigrate")
	}

	command := args[1]
	switch command {
	case "up", "automigrate":
		logger.Info("Starting database migration (Up)...")
		err := db.AutoMigrate(
			&user.User{},
			&audit.AuditTrail{},
			&blog.Blog{},
			&user.UserSocial{},
			&user.PasswordReset{},
			&user.EmailVerification{},
			&user.JWTBlacklist{},
		)
		if err != nil {
			logger.Fatal(err, "Database migration failed")
		}
		logger.Info("Database migration completed successfully")

	case "rollback":
		logger.Info("Starting database rollback (Down/Reset)...")
		err := db.Migrator().DropTable(
			&user.EmailVerification{},
			&user.PasswordReset{},
			&user.UserSocial{},
			&blog.Blog{},
			&audit.AuditTrail{},
			&user.User{},
			&user.JWTBlacklist{},
		)
		if err != nil {
			logger.Fatal(err, "Database rollback failed")
		}
		logger.Info("Database rollback completed successfully")

	default:
		log.Fatalf("Unknown migration command: %s. Use up, rollback, or automigrate", command)
	}
}
