package user

import (
	"zerobin-be/internal/middleware"
	"zerobin-be/internal/modules/audit"
	"zerobin-be/pkg/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	jwtService *auth.JWTService,
	enforcer auth.Enforcer,
	auditSvc audit.Service,
) {
	authLimit := middleware.RateLimit(1.0/12.0, 5)
	uploadLimit := middleware.RateLimit(1.0/20.0, 3)

	router.POST("/users/login", authLimit, middleware.AuditLog(auditSvc, "Auth", "Login"), handler.Login)
	router.POST("/users/login/oauth", authLimit, middleware.AuditLog(auditSvc, "Auth", "OAuthLogin"), handler.OAuthLogin)
	router.POST("/users/refresh", authLimit, handler.Refresh)
	router.POST("/users", authLimit, middleware.AuditLog(auditSvc, "User", "Register"), handler.CreateUser)
	router.POST("/users/forgot-password", authLimit, handler.ForgotPassword)
	router.POST("/users/reset-password", authLimit, middleware.AuditLog(auditSvc, "User", "ResetPassword"), handler.ResetPassword)
	router.POST("/users/verify-email", authLimit, middleware.AuditLog(auditSvc, "User", "VerifyEmail"), handler.VerifyEmail)
	router.POST("/users/verify-email/resend", authLimit, handler.ResendVerification)

	protected := router.Group("/users")
	protected.Use(middleware.Auth(jwtService))
	{
		protected.POST("/logout", middleware.AuditLog(auditSvc, "Auth", "Logout"), handler.Logout)
		protected.GET("", middleware.RBAC(enforcer), handler.GetUsers)

		protected.GET("/me", handler.GetMe)
		protected.PUT("/me", middleware.AuditLog(auditSvc, "User", "UpdateProfile"), handler.UpdateMe)
		protected.PUT("/me/password", middleware.AuditLog(auditSvc, "User", "ChangePassword"), handler.ChangePasswordMe)

		protected.GET("/:id", middleware.AuthorizeOwnership("id", middleware.UserSelfProvider), handler.GetUserByID)
		protected.PUT("/:id", middleware.AuthorizeOwnership("id", middleware.UserSelfProvider), middleware.AuditLog(auditSvc, "User", "Update"), handler.UpdateUser)
		protected.POST("/:id/avatar", middleware.AuthorizeOwnership("id", middleware.UserSelfProvider), uploadLimit, middleware.AuditLog(auditSvc, "User", "UploadAvatar"), handler.UploadAvatar)
		protected.PUT("/:id/password", middleware.AuthorizeOwnership("id", middleware.UserSelfProvider), middleware.AuditLog(auditSvc, "User", "ChangePassword"), handler.ChangePassword)

		protected.DELETE("/:id", middleware.RBAC(enforcer), middleware.AuditLog(auditSvc, "User", "Delete"), handler.DeleteUser)
	}

	admin := router.Group("/admin")
	admin.Use(middleware.Auth(jwtService), middleware.RBAC(enforcer))
	{
		admin.GET("/policies", handler.GetPolicies)
		admin.POST("/policies", middleware.AuditLog(auditSvc, "Policy", "Add"), handler.AddPolicy)
		admin.DELETE("/policies", middleware.AuditLog(auditSvc, "Policy", "Remove"), handler.RemovePolicy)
	}
}
