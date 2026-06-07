package middleware

import (
	"net/http"
	"strings"

	"cellar-backend/pkg/auth"
	"cellar-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	UserIDContextKey   = "userID"
	UserRoleContextKey = "userRole"
)

func Auth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header is required", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, http.StatusUnauthorized, "Authorization header must be Bearer token", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired authorization token", err.Error())
			c.Abort()
			return
		}

		// Inject details to request context
		c.Set(UserIDContextKey, claims.UserID)
		c.Set(UserRoleContextKey, claims.Role)
		c.Next()
	}
}

// RBAC checks if the current role is authorized to perform the action using Casbin
func RBAC(enforcer auth.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(UserRoleContextKey)
		if !exists {
			response.Error(c, http.StatusForbidden, "Access denied: missing role claims", nil)
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Error(c, http.StatusForbidden, "Access denied: invalid role format", nil)
			c.Abort()
			return
		}

		// Extract URL pattern and method
		obj := c.Request.URL.Path
		act := c.Request.Method

		allowed, err := enforcer.Enforce(roleStr, obj, act)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "RBAC enforcer error", err.Error())
			c.Abort()
			return
		}

		if !allowed {
			response.Error(c, http.StatusForbidden, "Access denied: unauthorized role privilege", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
