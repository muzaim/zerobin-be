package health

import (
	"context"
	"net/http"
	"time"

	"zerobin-be/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// HealthCheck check if app is running
// @Summary Check API Health
// @Description Returns if application server is running and database is reachable
// @Tags Health
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 503 {object} response.APIResponse
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "UP"
	sqlDB, err := h.db.DB()
	if err != nil {
		dbStatus = "DOWN"
	} else if err := sqlDB.PingContext(ctx); err != nil {
		dbStatus = "DOWN"
	}

	if dbStatus == "DOWN" {
		response.JSON(c, http.StatusServiceUnavailable, false, "Database connection is unhealthy", gin.H{
			"database": dbStatus,
		}, nil)
		return
	}

	response.Success(c, http.StatusOK, "Application is running", gin.H{
		"database": dbStatus,
	})
}
