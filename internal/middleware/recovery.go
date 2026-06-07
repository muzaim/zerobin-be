package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"cellar-backend/pkg/logger"
	"cellar-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Stack trace log
				logger.Error(fmt.Errorf("%v", err), "Panic recovered: %s\n%s", err, string(debug.Stack()))

				// Return standardized 500 API response
				response.Error(c, http.StatusInternalServerError, "Internal server error occurred", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
