package blog

import (
	"context"
	"errors"
	"testing"

	"zerobin-be/internal/modules/audit"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBlogRepository struct {
	mock.Mock
}

func (m *MockBlogRepository) Create(ctx context.Context, b *Blog) error {
	args := m.Called(ctx, b)
	b.ID = uuid.New()
	return args.Error(0)
}

func (m *MockBlogRepository) FindAll(ctx context.Context, category string, authorID string, search string) ([]Blog, error) {
	args := m.Called(ctx, category, authorID, search)
	return args.Get(0).([]Blog), args.Error(1)
}

func (m *MockBlogRepository) FindByID(ctx context.Context, id uuid.UUID) (*Blog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Blog), args.Error(1)
}

func (m *MockBlogRepository) FindBySlug(ctx context.Context, slug string) (*Blog, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Blog), args.Error(1)
}

func (m *MockBlogRepository) Update(ctx context.Context, b *Blog) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *MockBlogRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) Record(userID string, module string, action string, resourceID string, description string) error {
	args := m.Called(userID, module, action, resourceID, description)
	return args.Error(0)
}

func (m *MockAuditService) RecordTrail(trail *audit.AuditTrail) error {
	args := m.Called(trail)
	return args.Error(0)
}

func TestService_CreateBlog(t *testing.T) {
	repo := new(MockBlogRepository)
	auditSvc := new(MockAuditService)
	svc := NewService(repo, auditSvc)
	ctx := context.Background()
	authorID := uuid.New().String()

	t.Run("success without collision", func(t *testing.T) {
		req := &CreateBlogRequest{
			Title:    "My First Post",
			Desc:     "This is a description of my first post.",
			Category: "Education",
		}

		repo.On("FindBySlug", ctx, "my-first-post").Return(nil, errors.New("not found")).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*blog.Blog")).Return(nil).Once()
		repo.On("FindByID", ctx, mock.Anything).Return(&Blog{
			Title:    req.Title,
			Slug:     "my-first-post",
			Desc:     req.Desc,
			Category: req.Category,
		}, nil).Once()


		res, err := svc.Create(ctx, req, authorID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Title, res.Title)
		assert.Equal(t, "my-first-post", res.Slug)
		assert.Equal(t, req.Category, res.Category)

		repo.AssertExpectations(t)
		auditSvc.AssertExpectations(t)
	})

	t.Run("success with slug collision", func(t *testing.T) {
		req := &CreateBlogRequest{
			Title:    "My First Post",
			Desc:     "Another description.",
			Category: "Tips",
		}

		existingBlog := &Blog{Title: "Existing Post", Slug: "my-first-post"}

		repo.On("FindBySlug", ctx, "my-first-post").Return(existingBlog, nil).Once()
		repo.On("FindBySlug", ctx, "my-first-post-1").Return(nil, errors.New("not found")).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*blog.Blog")).Return(nil).Once()
		repo.On("FindByID", ctx, mock.Anything).Return(&Blog{
			Title:    req.Title,
			Slug:     "my-first-post-1",
			Desc:     req.Desc,
			Category: req.Category,
		}, nil).Once()


		res, err := svc.Create(ctx, req, authorID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "my-first-post-1", res.Slug)

		repo.AssertExpectations(t)
		auditSvc.AssertExpectations(t)
	})
}

func TestService_GetBlogs(t *testing.T) {
	repo := new(MockBlogRepository)
	auditSvc := new(MockAuditService)
	svc := NewService(repo, auditSvc)
	ctx := context.Background()

	t.Run("success list all", func(t *testing.T) {
		repo.On("FindAll", ctx, "", "", "").Return([]Blog{
			{Title: "Post One", Slug: "post-one", Category: "Education"},
			{Title: "Post Two", Slug: "post-two", Category: "Tips"},
		}, nil).Once()

		res, err := svc.GetAll(ctx, "", "", "")
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "post-one", res[0].Slug)
		repo.AssertExpectations(t)
	})
}

func TestService_UpdateBlog(t *testing.T) {
	repo := new(MockBlogRepository)
	auditSvc := new(MockAuditService)
	svc := NewService(repo, auditSvc)
	ctx := context.Background()
	blogID := uuid.New()
	operatorID := uuid.New().String()

	t.Run("success update title and slug", func(t *testing.T) {
		existing := &Blog{
			Title:    "Old Title",
			Slug:     "old-title",
			Category: "Education",
		}
		existing.ID = blogID

		req := &UpdateBlogRequest{
			Title: "New Title",
		}

		repo.On("FindByID", ctx, blogID).Return(existing, nil).Once()
		repo.On("FindBySlug", ctx, "new-title").Return(nil, errors.New("not found")).Once()
		repo.On("Update", ctx, mock.AnythingOfType("*blog.Blog")).Return(nil).Once()


		res, err := svc.Update(ctx, blogID, req, operatorID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "New Title", res.Title)
		assert.Equal(t, "new-title", res.Slug)

		repo.AssertExpectations(t)
		auditSvc.AssertExpectations(t)
	})
}
