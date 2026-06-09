package blog

import (
	"cellar-backend/internal/middleware"
	"cellar-backend/internal/modules/audit"
	"cellar-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	jwtService *auth.JWTService,
	enforcer auth.Enforcer,
	repo Repository,
	auditSvc audit.Service,
) {
	router.GET("/blogs", handler.GetBlogs)
	router.GET("/blogs/:id_or_slug", handler.GetBlog)

	protected := router.Group("/blogs")
	protected.Use(middleware.Auth(jwtService))
	{
		protected.POST("", middleware.RBAC(enforcer), middleware.AuditLog(auditSvc, "Blog", "Create"), handler.CreateBlog)
		protected.PUT("/:id", middleware.RBAC(enforcer), middleware.AuditLog(auditSvc, "Blog", "Update"), handler.UpdateBlog)
		protected.DELETE("/:id", middleware.RBAC(enforcer), middleware.AuditLog(auditSvc, "Blog", "Delete"), handler.DeleteBlog)
	}
}
