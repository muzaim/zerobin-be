package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"

	"zerobin-be/pkg/apperror"
	"zerobin-be/pkg/auth"

	"github.com/gin-gonic/gin"
)

const (
	UserIDContextKey   = "userID"
	UserRoleContextKey = "userRole"
)

type TokenBlacklistChecker interface {
	IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error)
}

var (
	blacklistChecker TokenBlacklistChecker
	blacklistMu      sync.RWMutex
)

func RegisterBlacklistChecker(checker TokenBlacklistChecker) {
	blacklistMu.Lock()
	defer blacklistMu.Unlock()
	blacklistChecker = checker
}

func getBlacklistChecker() TokenBlacklistChecker {
	blacklistMu.RLock()
	defer blacklistMu.RUnlock()
	return blacklistChecker
}

func Auth(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(apperror.New(apperror.TypeUnauthorized, "Authorization header is required", nil))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.Error(apperror.New(apperror.TypeUnauthorized, "Authorization header must be Bearer token", nil))
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.Error(apperror.New(apperror.TypeUnauthorized, "Invalid or expired authorization token", err))
			c.Abort()
			return
		}

		checker := getBlacklistChecker()
		if checker != nil {
			hash := sha256.Sum256([]byte(tokenString))
			tokenHash := hex.EncodeToString(hash[:])
			isBlacklisted, err := checker.IsTokenBlacklisted(c.Request.Context(), tokenHash)
			if err == nil && isBlacklisted {
				c.Error(apperror.New(apperror.TypeUnauthorized, "Token is invalidated or logged out", nil))
				c.Abort()
				return
			}
		}

		c.Set(UserIDContextKey, claims.UserID)
		c.Set(UserRoleContextKey, claims.Role)
		c.Next()
	}
}

func RBAC(enforcer auth.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(UserRoleContextKey)
		if !exists {
			c.Error(apperror.New(apperror.TypeForbidden, "Access denied: missing role claims", nil))
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.Error(apperror.New(apperror.TypeForbidden, "Access denied: invalid role format", nil))
			c.Abort()
			return
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		allowed, err := enforcer.Enforce(roleStr, obj, act)
		if err != nil {
			c.Error(apperror.New(apperror.TypeInternal, "RBAC enforcer error", err))
			c.Abort()
			return
		}

		if !allowed {
			c.Error(apperror.New(apperror.TypeForbidden, "Access denied: unauthorized role privilege", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
