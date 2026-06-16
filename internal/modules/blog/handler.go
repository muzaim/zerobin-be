package blog

import (
	"net/http"

	"zerobin-be/internal/middleware"
	"zerobin-be/pkg/response"
	"zerobin-be/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateBlog creates a new blog post
// @Summary Create Blog
// @Description Add a new blog post to the database
// @Tags Blogs
// @Accept json
// @Produce json
// @Param request body CreateBlogRequest true "Blog Create Body"
// @Security BearerAuth
// @Success 201 {object} response.APIResponse{data=BlogResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 422 {object} response.APIResponse
// @Router /blogs [post]
func (h *Handler) CreateBlog(c *gin.Context) {
	var req CreateBlogRequest
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

	blog, err := h.service.Create(c.Request.Context(), &req, operatorIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	c.Set(middleware.AuditResourceIDKey, blog.ID)
	response.Success(c, http.StatusCreated, "Blog post created successfully", blog)
}

// GetBlogs lists all blog posts
// @Summary List Blogs
// @Description Get all blog posts with optional filtering by category, user_id, or title/slug search
// @Tags Blogs
// @Produce json
// @Param category query string false "Filter by category (Education, Guides, Stories, Tips)"
// @Param user_id query string false "Filter by author User UUID"
// @Param search query string false "Search in title or slug"
// @Success 200 {object} response.APIResponse{data=[]BlogResponse}
// @Failure 500 {object} response.APIResponse
// @Router /blogs [get]
func (h *Handler) GetBlogs(c *gin.Context) {
	category := c.Query("category")
	userID := c.Query("user_id")
	search := c.Query("search")

	blogs, err := h.service.GetAll(c.Request.Context(), category, userID, search)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve blog posts", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Blog posts retrieved successfully", blogs)
}

// GetBlog retrieves a single blog post by UUID or Slug
// @Summary Get Blog By ID or Slug
// @Description Fetch a blog post's details using either its UUID or unique slug
// @Tags Blogs
// @Produce json
// @Param id_or_slug path string true "Blog ID or Slug"
// @Success 200 {object} response.APIResponse{data=BlogResponse}
// @Failure 404 {object} response.APIResponse
// @Router /blogs/{id_or_slug} [get]
func (h *Handler) GetBlog(c *gin.Context) {
	param := c.Param("id_or_slug")

	var blog *BlogResponse
	var err error

	if id, parseErr := uuid.Parse(param); parseErr == nil {
		blog, err = h.service.GetByID(c.Request.Context(), id)
	} else {
		blog, err = h.service.GetBySlug(c.Request.Context(), param)
	}

	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Blog post retrieved successfully", blog)
}

// UpdateBlog updates a blog post
// @Summary Update Blog
// @Description Modify an existing blog post by its UUID
// @Tags Blogs
// @Accept json
// @Produce json
// @Param id path string true "Blog UUID"
// @Param request body UpdateBlogRequest true "Blog Update Fields"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=BlogResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 422 {object} response.APIResponse
// @Router /blogs/{id} [put]
func (h *Handler) UpdateBlog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID parameter format", nil)
		return
	}

	var req UpdateBlogRequest
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

	blog, err := h.service.Update(c.Request.Context(), id, &req, operatorIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, http.StatusOK, "Blog post updated successfully", blog)
}

// DeleteBlog deletes a blog post
// @Summary Delete Blog
// @Description Delete a blog post by its UUID
// @Tags Blogs
// @Produce json
// @Param id path string true "Blog UUID"
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /blogs/{id} [delete]
func (h *Handler) DeleteBlog(c *gin.Context) {
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

	response.Success(c, http.StatusOK, "Blog post deleted successfully", nil)
}
