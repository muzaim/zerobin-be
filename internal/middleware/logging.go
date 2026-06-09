package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if raw != "" {
			path = path + "?" + raw
		}

		reqID, _ := c.Get(RequestIDKey)

		logEvent := log.Info().
			Str("request_id", reqID.(string)).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Str("latency", latency.String()).
			Str("ip", clientIP)

		if len(c.Errors) > 0 {
			logEvent.Interface("errors", c.Errors.JSON()).Msg("Request processed with errors")
		} else {
			logEvent.Msg("Request processed")
		}
	}
}
