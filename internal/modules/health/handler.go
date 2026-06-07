package health

import (
	"net/http"

	"cellar-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// HealthCheck check if app is running
// @Summary Check API Health
// @Description Returns if application server is running
// @Tags Health
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	response.Success(c, http.StatusOK, "Application is running", nil)
}
