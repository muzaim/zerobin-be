package middleware

import (
	"net/http"

	"cellar-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// OwnerProvider defines a function signature to retrieve the owner ID of a resource.
type OwnerProvider func(c *gin.Context, resourceID string) (string, error)

// AuthorizeOwnership is a reusable middleware enforcing resource ownership.
// If the user role is 'admin', they bypass ownership checks.
func AuthorizeOwnership(paramName string, provider OwnerProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(UserIDContextKey)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Unauthorized: missing authenticated user context", nil)
			c.Abort()
			return
		}

		role, roleExists := c.Get(UserRoleContextKey)
		if roleExists && role.(string) == "admin" {
			// Admin bypasses ownership verification
			c.Next()
			return
		}

		resourceID := c.Param(paramName)
		if resourceID == "" {
			response.Error(c, http.StatusBadRequest, "Bad Request: missing target resource identifier", nil)
			c.Abort()
			return
		}

		ownerID, err := provider(c, resourceID)
		if err != nil {
			response.Error(c, http.StatusForbidden, "Access Denied: resource not found or inaccessible", nil)
			c.Abort()
			return
		}

		if userID.(string) != ownerID {
			response.Error(c, http.StatusForbidden, "Forbidden: you do not have permission to access this resource", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserSelfProvider is a standard provider for entities where the resource identifier
// in the route parameter itself is the Owner ID (e.g. users module /users/:id).
func UserSelfProvider(c *gin.Context, resourceID string) (string, error) {
	return resourceID, nil
}

// IsOwner checks if the current context user owns the resource (non-middleware helper version)
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
