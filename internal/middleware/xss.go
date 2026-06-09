package middleware

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func XSS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !jsonRequest(contentType) {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		var generic interface{}
		if err := json.Unmarshal(body, &generic); err != nil {
			c.Next()
			return
		}

		sanitized := sanitizeInterface(generic)

		sanitizedBody, err := json.Marshal(sanitized)
		if err != nil {
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBody))
		c.Request.ContentLength = int64(len(sanitizedBody))
		c.Next()
	}
}

func jsonRequest(contentType string) bool {
	return (len(contentType) >= 16 && contentType[:16] == "application/json") ||
		(len(contentType) > 16 && contentType[16] == ';')
}

func sanitizeInterface(val interface{}) interface{} {
	switch v := val.(type) {
	case string:
		return html.EscapeString(v)
	case map[string]interface{}:
		sanitizedMap := make(map[string]interface{})
		for k, val := range v {
			sanitizedMap[k] = sanitizeInterface(val)
		}
		return sanitizedMap
	case []interface{}:
		sanitizedSlice := make([]interface{}, len(v))
		for i, val := range v {
			sanitizedSlice[i] = sanitizeInterface(val)
		}
		return sanitizedSlice
	default:
		return v
	}
}
