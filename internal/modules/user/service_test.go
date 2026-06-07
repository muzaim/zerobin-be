package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"cellar-backend/configs"
	"cellar-backend/internal/entities"
	"cellar-backend/pkg/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository mocks the GORM repository layer
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	u.ID = uuid.New() // simulate database assigning UUID
	return args.Error(0)
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockAuditService mocks the Audit Trail business logic layer
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) Record(userID string, module string, action string, description string) error {
	args := m.Called(userID, module, action, description)
	return args.Error(0)
}

func TestService_CreateUser(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})

	svc := NewService(repo, auditSvc, jwtSvc)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := &CreateUserRequest{
			FullName: "John Doe",
			Email:    "john@example.com",
			Password: "securepassword",
			Role:     "user",
		}

		repo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("record not found")).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()
		auditSvc.On("Record", "admin-id", "User", "Create", mock.AnythingOfType("string")).Return(nil).Once()

		res, err := svc.Create(ctx, req, "admin-id")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.FullName, res.FullName)
		assert.Equal(t, req.Email, res.Email)
		assert.Equal(t, "user", res.Role)

		repo.AssertExpectations(t)
		auditSvc.AssertExpectations(t)
	})

	t.Run("duplicate email error", func(t *testing.T) {
		req := &CreateUserRequest{
			FullName: "John Doe",
			Email:    "duplicate@example.com",
			Password: "securepassword",
		}

		existingUser := &User{
			FullName: "Existing",
			Email:    "duplicate@example.com",
		}

		repo.On("FindByEmail", ctx, req.Email).Return(existingUser, nil).Once()

		res, err := svc.Create(ctx, req, "admin-id")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "email address is already in use", err.Error())

		repo.AssertExpectations(t)
	})
}

func TestService_Login(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})

	svc := NewService(repo, auditSvc, jwtSvc)
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	dbUser := &User{
		Base: entities.Base{
			ID: uuid.New(),
		},
		FullName: "Login User",
		Email:    "login@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}

	t.Run("success", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "login@example.com",
			Password: "correctpassword",
		}

		repo.On("FindByEmail", ctx, req.Email).Return(dbUser, nil).Once()
		auditSvc.On("Record", dbUser.ID.String(), "Auth", "Login", mock.AnythingOfType("string")).Return(nil).Once()

		res, err := svc.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.Equal(t, dbUser.Email, res.User.Email)

		repo.AssertExpectations(t)
		auditSvc.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "login@example.com",
			Password: "wrongpassword",
		}

		repo.On("FindByEmail", ctx, req.Email).Return(dbUser, nil).Once()

		res, err := svc.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "invalid email or password", err.Error())

		repo.AssertExpectations(t)
	})
}
