package user

import (
	"net/http"

	"cellar-backend/internal/middleware"
	"cellar-backend/pkg/response"
	"cellar-backend/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
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
	response.Success(c, http.StatusOK, "Login successful", res)
}
