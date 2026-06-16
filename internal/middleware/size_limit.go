package middleware

import (
	"net/http"

	"zerobin-be/pkg/response"

	"github.com/gin-gonic/gin"
)

func RequestSizeLimiter(defaultLimit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := defaultLimit
		if c.FullPath() == "/api/v1/users/:id/avatar" {
			limit = 5 * 1024 * 1024
		}

		if c.Request.ContentLength > limit {
			response.Error(c, http.StatusRequestEntityTooLarge, "Request payload exceeds size limit", nil)
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
