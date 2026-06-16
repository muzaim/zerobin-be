package middleware

import (
	"zerobin-be/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type OwnerProvider func(c *gin.Context, resourceID string) (string, error)

func AuthorizeOwnership(paramName string, provider OwnerProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(UserIDContextKey)
		if !exists {
			c.Error(apperror.New(apperror.TypeUnauthorized, "Unauthorized: missing authenticated user context", nil))
			c.Abort()
			return
		}

		role, roleExists := c.Get(UserRoleContextKey)
		if roleExists && role.(string) == "admin" {
			c.Next()
			return
		}

		resourceID := c.Param(paramName)
		if resourceID == "" {
			c.Error(apperror.New(apperror.TypeBadRequest, "Bad Request: missing target resource identifier", nil))
			c.Abort()
			return
		}

		ownerID, err := provider(c, resourceID)
		if err != nil {
			c.Error(apperror.New(apperror.TypeForbidden, "Access Denied: resource not found or inaccessible", err))
			c.Abort()
			return
		}

		if userID.(string) != ownerID {
			c.Error(apperror.New(apperror.TypeForbidden, "Forbidden: you do not have permission to access this resource", nil))
			c.Abort()
			return
		}

		c.Next()
	}
}

func UserSelfProvider(c *gin.Context, resourceID string) (string, error) {
	return resourceID, nil
}

func IsOwner(c *gin.Context, ownerID string) bool {
	userID, exists := c.Get(UserIDContextKey)
	if !exists {
		return false
	}

	role, roleExists := c.Get(UserRoleContextKey)
	if roleExists && role.(string) == "admin" {
		return true
	}

	return userID.(string) == ownerID
}
