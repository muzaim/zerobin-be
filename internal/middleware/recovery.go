package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"zerobin-be/pkg/logger"
	"zerobin-be/pkg/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(fmt.Errorf("%v", err), "Panic recovered: %s\n%s", err, string(debug.Stack()))

				response.Error(c, http.StatusInternalServerError, "Internal server error occurred", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
