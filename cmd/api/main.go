package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zerobin-be/configs"
	_ "zerobin-be/docs"
	"zerobin-be/internal/middleware"
	"zerobin-be/internal/modules/asset"
	"zerobin-be/internal/modules/audit"
	"zerobin-be/internal/modules/health"
	"zerobin-be/internal/modules/blog"
	"zerobin-be/internal/modules/user"
	"zerobin-be/internal/worker"
	"zerobin-be/pkg/auth"
	"zerobin-be/pkg/database"
	"zerobin-be/pkg/email"
	"zerobin-be/pkg/logger"
	"zerobin-be/pkg/storage"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Golang SaaS Starter Kit API
// @version 1.0
// @description Production-ready Golang SaaS backend starter template.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer " followed by your JWT token.
func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	logger.InitLogger(cfg.App.Env, cfg.App.LogLevel)

	db, err := database.InitDB(&cfg.DB, cfg.App.Env)
	if err != nil {
		logger.Fatal(err, "Failed to connect to database")
	}

	if cfg.DB.AutoMigrate {
		logger.Info("Auto-migration enabled. Running migrations...")
		err = db.AutoMigrate(
			&asset.Asset{},
			&user.User{},
			&audit.AuditTrail{},
			&blog.Blog{},
			&user.UserSocial{},
			&user.PasswordReset{},
			&user.EmailVerification{},
			&user.JWTBlacklist{},
		)
		if err != nil {
			logger.Fatal(err, "Auto-migration failed")
		}
		logger.Info("Auto-migration completed successfully")
	}

	jwtService := auth.NewJWTService(&cfg.JWT)

	enforcer, err := auth.NewCasbinEnforcer("configs/rbac_model.conf", db)
	if err != nil {
		logger.Fatal(err, "Failed to load Casbin RBAC policies")
	}

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {}

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.Logging())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(&cfg.CORS))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestSizeLimiter(2 * 1024 * 1024))
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.XSS())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	auditRepo := audit.NewRepository(db)
	auditServ := audit.NewService(auditRepo)

	storageServ, err := storage.NewMinIOStorage(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKeyID,
		cfg.MinIO.SecretAccessKey,
		cfg.MinIO.BucketName,
		cfg.MinIO.PublicURL,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		logger.Fatal(err, "Failed to initialize MinIO storage")
	}

	emailServ := email.NewEmailService(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.Sender)

	txManager := database.NewTransactionManager(db)
	assetRepo := asset.NewRepository(db)
	assetServ := asset.NewService(assetRepo, storageServ)
	userRepo := user.NewRepository(db)
	middleware.RegisterBlacklistChecker(userRepo)
	userServ := user.NewService(userRepo, auditServ, jwtService, storageServ, emailServ, txManager, assetServ)

	blogRepo := blog.NewRepository(db)
	blogServ := blog.NewService(blogRepo, auditServ)

	healthHandler := health.NewHandler(db)
	userHandler := user.NewHandler(userServ, enforcer)
	blogHandler := blog.NewHandler(blogServ)

	health.RegisterRoutes(v1, healthHandler)
	user.RegisterRoutes(v1, userHandler, jwtService, enforcer, auditServ)
	blog.RegisterRoutes(v1, blogHandler, jwtService, enforcer, blogRepo, auditServ)

	cleanupWorker := worker.NewCleanupWorker(db)
	cleanupWorker.Start(1 * time.Hour)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.App.Port),
		Handler: r,
	}

	go func() {
		logger.Info("Starting HTTP server on port %s...", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err, "HTTP server listening failure")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down HTTP server...")

	cleanupWorker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(err, "Server forced to shutdown")
	}

	logger.Info("Server exited gracefully")
}
