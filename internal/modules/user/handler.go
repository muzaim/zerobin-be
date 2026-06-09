package user

import (
	"io"
	"net/http"
	"strings"

	"cellar-backend/internal/middleware"
	"cellar-backend/pkg/apperror"
	"cellar-backend/pkg/auth"
	"cellar-backend/pkg/response"
	"cellar-backend/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service  Service
	enforcer auth.Enforcer
}

func NewHandler(service Service, enforcer auth.Enforcer) *Handler {
	return &Handler{
		service:  service,
		enforcer: enforcer,
	}
}

// GetUsers lists all users
// @Summary List Users
// @Description Get all registered users in the database
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]UserResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	users, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve users list", err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Users retrieved successfully", users)
}

// GetUserByID retrieves a specific user by UUID
// @Summary Get User By ID
// @Description Fetch user details using UUID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=UserResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID parameter format", nil)
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User retrieved successfully", user)
}

// CreateUser creates a new user
// @Summary Create User
// @Description Add a new user to the database
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User Registration Body"
// @Success 201 {object} response.APIResponse{data=UserResponse}
// @Failure 400 {object} response.APIResponse
// @Router /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}

	operatorID, _ := c.Get(middleware.UserIDContextKey)
	operatorIDStr, _ := operatorID.(string) // Could be empty string if anonymous register

	user, err := h.service.Create(c.Request.Context(), &req, operatorIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	c.Set(middleware.AuditResourceIDKey, user.ID)
	response.Success(c, http.StatusCreated, "User created successfully", user)
}

// UpdateUser updates user attributes
// @Summary Update User
// @Description Modify existing user profiles by UUID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "Update Fields Body"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=UserResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID parameter format", nil)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}

	operatorID, _ := c.Get(middleware.UserIDContextKey)
	operatorIDStr, _ := operatorID.(string)

	user, err := h.service.Update(c.Request.Context(), id, &req, operatorIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User updated successfully", user)
}

// DeleteUser deletes a user
// @Summary Delete User
// @Description Soft delete a user by UUID
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID parameter format", nil)
		return
	}

	operatorID, _ := c.Get(middleware.UserIDContextKey)
	operatorIDStr, _ := operatorID.(string)

	err = h.service.Delete(c.Request.Context(), id, operatorIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "User deleted successfully", nil)
}

// Login authenticates a user and returns access/refresh tokens
// @Summary Login User
// @Description Authenticate using email and password to retrieve access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Credentials"
// @Success 200 {object} response.APIResponse{data=LoginResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}

	res, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}
	c.Set(middleware.UserIDContextKey, res.User.ID)
	c.Set(middleware.AuditResourceIDKey, res.User.ID)
	response.Success(c, http.StatusOK, "Login successful", res)
}

// OAuthLogin authenticates a user via social login (Google/Apple) and returns access/refresh tokens
// @Summary OAuth Login User
// @Description Authenticate using social provider ID Token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body OAuthLoginRequest true "OAuth Credentials"
// @Success 200 {object} response.APIResponse{data=LoginResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/login/oauth [post]
func (h *Handler) OAuthLogin(c *gin.Context) {
	var req OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}

	res, err := h.service.OAuthLogin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}
	c.Set(middleware.UserIDContextKey, res.User.ID)
	c.Set(middleware.AuditResourceIDKey, res.User.ID)
	response.Success(c, http.StatusOK, "OAuth Login successful", res)
}

// @Summary Upload User Avatar
// @Description Upload profile picture and update avatar_url for a user
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User ID"
// @Param avatar formData file true "Avatar Image File"
// @Success 200 {object} response.APIResponse{data=UserResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /users/{id}/avatar [post]
func (h *Handler) UploadAvatar(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Missing avatar file in request", nil)
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 {
		response.Error(c, http.StatusBadRequest, "File size exceeds limit of 5MB", nil)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		if _, err := file.Read(buf); err == nil {
			contentType = http.DetectContentType(buf)
			_, _ = file.Seek(0, io.SeekStart)
		}
	}

	res, err := h.service.UpdateAvatar(c.Request.Context(), userID, file, header.Size, contentType)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Avatar uploaded successfully", res)
}

// @Summary Refresh Access Token
// @Description Exchange a valid refresh token for a new access token and rotated refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} response.APIResponse{data=LoginResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 422 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /users/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}

	res, err := h.service.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", res)
}

// ChangePassword changes user password
// @Summary Change Password
// @Description Update user password by verifying the current password
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body ChangePasswordRequest true "Change Password Request"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/{id}/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID parameter format", nil)
		return
	}
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}
	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), id, &req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Password changed successfully", nil)
}

// ForgotPassword requests a password reset link
// @Summary Forgot Password
// @Description Request a password reset link to be sent to user email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /users/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}
	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}
	if err := h.service.ForgotPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process forgot password request", err.Error())
		return
	}
	response.Success(c, http.StatusOK, "If the email is registered, a password reset link has been sent", nil)
}

// ResetPassword resets user password using a token
// @Summary Reset Password
// @Description Reset password using a valid reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /users/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}
	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Password has been reset successfully", nil)
}

// VerifyEmail verifies user email
// @Summary Verify Email
// @Description Verify a registered user using the activation token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Verify Email Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /users/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}
	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}
	if err := h.service.VerifyEmail(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "Email verified successfully", nil)
}

// ResendVerification resends the activation email
// @Summary Resend Verification Email
// @Description Request a new verification link to be sent to user email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResendVerificationRequest true "Resend Verification Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /users/verify-email/resend [post]
func (h *Handler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON body format", err.Error())
		return
	}
	if errs := validator.ValidateStruct(&req); errs != nil {
		response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)
		return
	}
	if err := h.service.ResendVerification(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, http.StatusOK, "If the email is unverified, a new verification link has been sent", nil)
}

// @Summary Get Current User Profile
// @Description Fetch details of the currently authenticated session
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=UserResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.Error(apperror.New(apperror.TypeUnauthorized, "Missing authenticated user context", nil))
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.Error(apperror.New(apperror.TypeUnauthorized, "Invalid user ID context format", nil))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid user ID UUID format", err))
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.Error(apperror.New(apperror.TypeNotFound, err.Error(), err))
		return
	}
	response.Success(c, http.StatusOK, "Profile retrieved successfully", user)
}

// @Summary Update Current User Profile
// @Description Modify fields of the currently authenticated profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateUserRequest true "Update Profile Fields"
// @Success 200 {object} response.APIResponse{data=UserResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/me [put]
func (h *Handler) UpdateMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.Error(apperror.New(apperror.TypeUnauthorized, "Missing authenticated user context", nil))
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.Error(apperror.New(apperror.TypeUnauthorized, "Invalid user ID context format", nil))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid user ID UUID format", err))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid JSON body format", err))
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		c.Error(apperror.NewWithDetail(apperror.TypeValidation, "Validation failed", errs, nil))
		return
	}

	roleVal, roleExists := c.Get(middleware.UserRoleContextKey)
	roleStr, _ := roleVal.(string)
	if !roleExists || roleStr != "admin" {
		req.Role = ""
	}

	user, err := h.service.Update(c.Request.Context(), userID, &req, userIDStr)
	if err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, err.Error(), err))
		return
	}
	response.Success(c, http.StatusOK, "Profile updated successfully", user)
}

// @Summary Change Current User Password
// @Description Update authenticated user password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Change Password Body"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/me/password [put]
func (h *Handler) ChangePasswordMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.UserIDContextKey)
	if !exists {
		c.Error(apperror.New(apperror.TypeUnauthorized, "Missing authenticated user context", nil))
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.Error(apperror.New(apperror.TypeUnauthorized, "Invalid user ID context format", nil))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid user ID UUID format", err))
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid JSON body format", err))
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		c.Error(apperror.NewWithDetail(apperror.TypeValidation, "Validation failed", errs, nil))
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, err.Error(), err))
		return
	}
	response.Success(c, http.StatusOK, "Password changed successfully", nil)
}

// @Summary Get Casbin Policies
// @Description Fetch all policy and grouping rules from Casbin DB
// @Tags Admin Policies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]PolicyRuleResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Router /admin/policies [get]
func (h *Handler) GetPolicies(c *gin.Context) {
	policies, err := h.enforcer.GetPolicy()
	if err != nil {
		c.Error(apperror.New(apperror.TypeInternal, "Failed to retrieve policy rules", err))
		return
	}
	groupings, err := h.enforcer.GetGroupingPolicy()
	if err != nil {
		c.Error(apperror.New(apperror.TypeInternal, "Failed to retrieve grouping rules", err))
		return
	}

	var res []PolicyRuleResponse
	for _, p := range policies {
		res = append(res, PolicyRuleResponse{
			PType: "p",
			Rule:  p,
		})
	}
	for _, g := range groupings {
		res = append(res, PolicyRuleResponse{
			PType: "g",
			Rule:  g,
		})
	}

	response.Success(c, http.StatusOK, "Policies retrieved successfully", res)
}

// @Summary Add Casbin Policy
// @Description Add a new policy or grouping rule to Casbin DB
// @Tags Admin Policies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PolicyRequest true "Policy Rule Details"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Router /admin/policies [post]
func (h *Handler) AddPolicy(c *gin.Context) {
	var req PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid JSON body format", err))
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		c.Error(apperror.NewWithDetail(apperror.TypeValidation, "Validation failed", errs, nil))
		return
	}

	if req.PType == "p" {
		if len(req.Rule) != 3 {
			c.Error(apperror.New(apperror.TypeBadRequest, "Policy rule must have exactly 3 elements: [sub, obj, act]", nil))
			return
		}
		added, err := h.enforcer.AddPolicy(req.Rule[0], req.Rule[1], req.Rule[2])
		if err != nil {
			c.Error(apperror.New(apperror.TypeInternal, "Failed to add policy rule", err))
			return
		}
		if !added {
			c.Error(apperror.New(apperror.TypeConflict, "Policy rule already exists", nil))
			return
		}
	} else {
		if len(req.Rule) != 2 {
			c.Error(apperror.New(apperror.TypeBadRequest, "Grouping rule must have exactly 2 elements: [sub, group]", nil))
			return
		}
		added, err := h.enforcer.AddGroupingPolicy(req.Rule[0], req.Rule[1])
		if err != nil {
			c.Error(apperror.New(apperror.TypeInternal, "Failed to add grouping policy rule", err))
			return
		}
		if !added {
			c.Error(apperror.New(apperror.TypeConflict, "Grouping policy rule already exists", nil))
			return
		}
	}

	response.Success(c, http.StatusOK, "Policy rule added successfully", nil)
}

// @Summary Remove Casbin Policy
// @Description Remove an existing policy or grouping rule from Casbin DB
// @Tags Admin Policies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PolicyRequest true "Policy Rule Details"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Router /admin/policies [delete]
func (h *Handler) RemovePolicy(c *gin.Context) {
	var req PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.New(apperror.TypeBadRequest, "Invalid JSON body format", err))
		return
	}

	if errs := validator.ValidateStruct(&req); errs != nil {
		c.Error(apperror.NewWithDetail(apperror.TypeValidation, "Validation failed", errs, nil))
		return
	}

	if req.PType == "p" {
		if len(req.Rule) != 3 {
			c.Error(apperror.New(apperror.TypeBadRequest, "Policy rule must have exactly 3 elements: [sub, obj, act]", nil))
			return
		}
		removed, err := h.enforcer.RemovePolicy(req.Rule[0], req.Rule[1], req.Rule[2])
		if err != nil {
			c.Error(apperror.New(apperror.TypeInternal, "Failed to remove policy rule", err))
			return
		}
		if !removed {
			c.Error(apperror.New(apperror.TypeNotFound, "Policy rule not found", nil))
			return
		}
	} else {
		if len(req.Rule) != 2 {
			c.Error(apperror.New(apperror.TypeBadRequest, "Grouping rule must have exactly 2 elements: [sub, group]", nil))
			return
		}
		removed, err := h.enforcer.RemoveGroupingPolicy(req.Rule[0], req.Rule[1])
		if err != nil {
			c.Error(apperror.New(apperror.TypeInternal, "Failed to remove grouping policy rule", err))
			return
		}
		if !removed {
			c.Error(apperror.New(apperror.TypeNotFound, "Grouping policy rule not found", nil))
			return
		}
	}

	response.Success(c, http.StatusOK, "Policy rule removed successfully", nil)
}

// Logout invalidates the user's active session token
// @Summary Logout User
// @Description Invalidate the active Bearer token by adding it to the blacklist
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Error(c, http.StatusBadRequest, "Invalid Authorization header", nil)
		return
	}

	token := parts[1]
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Successfully logged out", nil)
}


