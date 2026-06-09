package worker

import (
	"time"

	"cellar-backend/pkg/logger"

	"gorm.io/gorm"
)

type CleanupWorker struct {
	db     *gorm.DB
	ticker *time.Ticker
	stop   chan struct{}
}

func NewCleanupWorker(db *gorm.DB) *CleanupWorker {
	return &CleanupWorker{
		db:   db,
		stop: make(chan struct{}),
	}
}

func (w *CleanupWorker) Start(interval time.Duration) {
	w.ticker = time.NewTicker(interval)
	go func() {
		logger.Info("Starting background cleanup worker...")
		w.runCleanup()

		for {
			select {
			case <-w.ticker.C:
				w.runCleanup()
			case <-w.stop:
				logger.Info("Stopping background cleanup worker...")
				w.ticker.Stop()
				return
			}
		}
	}()
}

func (w *CleanupWorker) Stop() {
	close(w.stop)
}

func (w *CleanupWorker) runCleanup() {
	now := time.Now()
	logger.Info("Running expired tokens database cleanup...")

	if err := w.db.Exec("DELETE FROM sys_jwt_blacklist WHERE expires_at < ?", now).Error; err != nil {
		logger.Error(err, "Failed to clean expired blacklisted JWTs")
	}

	if err := w.db.Exec("DELETE FROM sys_password_resets WHERE expires_at < ?", now).Error; err != nil {
		logger.Error(err, "Failed to clean expired password resets")
	}

	if err := w.db.Exec("DELETE FROM sys_email_verifications WHERE expires_at < ?", now).Error; err != nil {
		logger.Error(err, "Failed to clean expired email verifications")
	}

	logger.Info("Database cleanup completed successfully")
}
