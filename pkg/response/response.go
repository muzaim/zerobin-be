package response

import (
	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// JSON sends a custom structured JSON response
func JSON(c *gin.Context, statusCode int, success bool, message string, data interface{}, errors interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Errors:  errors,
	})
}

// Success sends a standard successful JSON response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	JSON(c, statusCode, true, message, data, nil)
}

// Error sends a standard error JSON response
func Error(c *gin.Context, statusCode int, message string, errors interface{}) {
	JSON(c, statusCode, false, message, nil, errors)
}
