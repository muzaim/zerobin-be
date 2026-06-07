package user

import (
	"cellar-backend/internal/middleware"
	"cellar-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	jwtService *auth.JWTService,
	enforcer auth.Enforcer,
) {
	// Public routes
	router.POST("/users/login", handler.Login)
	router.POST("/users", handler.CreateUser) // Can be public for self-registration

	// Protected routes
	protected := router.Group("/users")
	protected.Use(middleware.Auth(jwtService))
	{
		// Requires admin or moderator (enforced via Casbin)
		protected.GET("", middleware.RBAC(enforcer), handler.GetUsers)

		// Self-access or Admin/Mod (enforced via ownership authorization)
		// UserSelfProvider assumes the resource ID in parameter 'id' is the user's owner ID.
		protected.GET("/:id", middleware.AuthorizeOwnership("id", middleware.UserSelfProvider), handler.GetUserByID)
		protected.PUT("/:id", middleware.AuthorizeOwnership("id", middleware.UserSelfProvider), handler.UpdateUser)

		// Only admin can delete users (enforced via Casbin)
		protected.DELETE("/:id", middleware.RBAC(enforcer), handler.DeleteUser)
	}
}
