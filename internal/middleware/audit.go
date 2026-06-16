package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"zerobin-be/internal/modules/audit"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AuditResourceIDKey  = "AuditResourceID"
	AuditDescriptionKey = "AuditDescription"
)

func AuditLog(auditSvc audit.Service, module string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload []byte
		if c.Request.Body != nil && (c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete) {
			var err error
			payload, err = io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(payload))
			}
		}

		sanitizedPayload := sanitizePayload(payload)
		startTime := time.Now()

		c.Next()

		var userIDUUID *uuid.UUID
		userIDVal, exists := c.Get(UserIDContextKey)
		if exists {
			if userIDStr, ok := userIDVal.(string); ok {
				if parsed, err := uuid.Parse(userIDStr); err == nil {
					userIDUUID = &parsed
				}
			}
		}

		resourceID := ""
		if resIDVal, ok := c.Get(AuditResourceIDKey); ok {
			resourceID, _ = resIDVal.(string)
		} else {
			resourceID = c.Param("id")
			if resourceID == "" {
				resourceID = c.Param("id_or_slug")
			}
		}

		statusCode := c.Writer.Status()
		status := "success"
		errMsg := ""
		if len(c.Errors) > 0 {
			status = "failure"
			errMsg = c.Errors.Last().Error()
		} else if statusCode >= 400 {
			status = "failure"
			errMsg = http.StatusText(statusCode)
		}

		description := ""
		if descVal, ok := c.Get(AuditDescriptionKey); ok {
			description, _ = descVal.(string)
		} else {
			description = strings.TrimSpace(action + " " + module)
			if resourceID != "" {
				description += " (" + resourceID + ")"
			}
		}

		reqID, _ := c.Get(RequestIDKey)
		reqIDStr, _ := reqID.(string)
		executionTime := time.Since(startTime).Milliseconds()

		trail := &audit.AuditTrail{
			UserID:        userIDUUID,
			Module:        module,
			Action:        action,
			ResourceID:    resourceID,
			Description:   description,
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			RequestID:     reqIDStr,
			Payload:       sanitizedPayload,
			Status:        status,
			StatusCode:    statusCode,
			ExecutionTime: executionTime,
			ErrorMessage:  errMsg,
			CreatedAt:     startTime,
		}

		_ = auditSvc.RecordTrail(trail)
	}
}

func sanitizePayload(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return string(payload)
	}

	sensitiveFields := []string{"password", "current_password", "new_password", "token", "id_token", "access_token", "refresh_token"}

	var sanitizeMap func(map[string]interface{})
	sanitizeMap = func(m map[string]interface{}) {
		for k, v := range m {
			isSensitive := false
			for _, field := range sensitiveFields {
				if strings.ToLower(k) == field {
					isSensitive = true
					break
				}
			}
			if isSensitive {
				m[k] = "[REDACTED]"
			} else if nestedMap, ok := v.(map[string]interface{}); ok {
				sanitizeMap(nestedMap)
			}
		}
	}

	sanitizeMap(data)

	sanitizedBytes, err := json.Marshal(data)
	if err != nil {
		return string(payload)
	}

	return string(sanitizedBytes)
}
