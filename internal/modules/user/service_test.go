package user

import (
	"context"
	"errors"
	"testing"
	"time"
	"io"
	"strings"

	"zerobin-be/configs"
	"zerobin-be/internal/entities"
	"zerobin-be/internal/modules/asset"
	"zerobin-be/internal/modules/audit"
	"zerobin-be/pkg/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	u.ID = uuid.New()
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

func (m *MockUserRepository) FindSocial(ctx context.Context, provider string, providerID string) (*UserSocial, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserSocial), args.Error(1)
}

func (m *MockUserRepository) CreateSocial(ctx context.Context, social *UserSocial) error {
	args := m.Called(ctx, social)
	return args.Error(0)
}

func (m *MockUserRepository) CreateResetToken(ctx context.Context, reset *PasswordReset) error {
	args := m.Called(ctx, reset)
	return args.Error(0)
}

func (m *MockUserRepository) FindResetToken(ctx context.Context, token string) (*PasswordReset, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PasswordReset), args.Error(1)
}

func (m *MockUserRepository) DeleteResetToken(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserRepository) CreateVerificationToken(ctx context.Context, verify *EmailVerification) error {
	args := m.Called(ctx, verify)
	return args.Error(0)
}

func (m *MockUserRepository) FindVerificationToken(ctx context.Context, token string) (*EmailVerification, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EmailVerification), args.Error(1)
}

func (m *MockUserRepository) DeleteVerificationToken(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserRepository) AddToBlacklist(ctx context.Context, blacklist *JWTBlacklist) error {
	args := m.Called(ctx, blacklist)
	return args.Error(0)
}

func (m *MockUserRepository) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
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

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) UploadAvatar(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) (string, error) {
	args := m.Called(ctx, filename, reader, size, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) UploadFile(ctx context.Context, folder string, filename string, reader io.Reader, size int64, contentType string) (string, error) {
	args := m.Called(ctx, folder, filename, reader, size, contentType)
	return args.String(0), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendHTML(to string, subject string, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendPlain(to string, subject string, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendTemplate(to string, subject string, templatePath string, data interface{}) error {
	args := m.Called(to, subject, templatePath, data)
	return args.Error(0)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type MockAssetService struct {
	mock.Mock
}

func (m *MockAssetService) Upload(ctx context.Context, filename string, folder string, reader io.Reader, size int64, contentType string, uploadedByID *uuid.UUID) (*asset.Asset, error) {
	args := m.Called(ctx, filename, folder, reader, size, contentType, uploadedByID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetService) GetByID(ctx context.Context, id uuid.UUID) (*asset.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
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
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)

	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
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
		repo.On("CreateVerificationToken", ctx, mock.AnythingOfType("*user.EmailVerification")).Return(nil).Once()
		emailSvc.On("SendTemplate", req.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		res, err := svc.Create(ctx, req, "admin-id")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.FullName, res.FullName)
		assert.Equal(t, req.Email, res.Email)
		assert.Equal(t, "user", res.Role)

		repo.AssertExpectations(t)
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
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)

	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	hashedPassword, _ := auth.HashPassword("correctpassword")
	dbUser := &User{
		Base: entities.Base{
			ID: uuid.New(),
		},
		FullName:   "Login User",
		Email:      "login@example.com",
		Password:   string(hashedPassword),
		Role:       "user",
		IsVerified: true,
	}

	t.Run("success", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "login@example.com",
			Password: "correctpassword",
		}

		repo.On("FindByEmail", ctx, req.Email).Return(dbUser, nil).Once()

		res, err := svc.Login(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.Equal(t, dbUser.Email, res.User.Email)

		repo.AssertExpectations(t)
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

	t.Run("unverified email error", func(t *testing.T) {
		unverifiedUser := &User{
			Base: entities.Base{
				ID: uuid.New(),
			},
			FullName:   "Unverified User",
			Email:      "unverified@example.com",
			Password:   string(hashedPassword),
			Role:       "user",
			IsVerified: false,
		}
		req := &LoginRequest{
			Email:    "unverified@example.com",
			Password: "correctpassword",
		}

		repo.On("FindByEmail", ctx, req.Email).Return(unverifiedUser, nil).Once()

		res, err := svc.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "please verify your email address first", err.Error())

		repo.AssertExpectations(t)
	})
}

func TestService_OAuthLogin(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)

	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	t.Run("success with existing social login", func(t *testing.T) {
		req := &OAuthLoginRequest{
			Provider:   "google",
			IDToken:    "mock-token",
			Email:      "oauth@example.com",
			FullName:   "Google User",
			ProviderID: "google-uid-123",
		}

		userSocial := &UserSocial{
			UserID:     uuid.New(),
			Provider:   req.Provider,
			ProviderID: req.ProviderID,
		}

		dbUser := &User{
			FullName: req.FullName,
			Email:    req.Email,
			Role:     "user",
		}
		dbUser.ID = userSocial.UserID

		repo.On("FindSocial", ctx, req.Provider, req.ProviderID).Return(userSocial, nil).Once()
		repo.On("FindByID", ctx, userSocial.UserID).Return(dbUser, nil).Once()

		res, err := svc.OAuthLogin(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.Equal(t, dbUser.Email, res.User.Email)

		repo.AssertExpectations(t)
	})

	t.Run("success link social to existing user email", func(t *testing.T) {
		req := &OAuthLoginRequest{
			Provider:   "google",
			IDToken:    "mock-token",
			Email:      "link@example.com",
			FullName:   "Linked User",
			ProviderID: "google-uid-456",
		}

		dbUser := &User{
			FullName: req.FullName,
			Email:    req.Email,
			Role:     "user",
		}
		dbUser.ID = uuid.New()

		repo.On("FindSocial", ctx, req.Provider, req.ProviderID).Return(nil, errors.New("not found")).Once()
		repo.On("FindByEmail", ctx, req.Email).Return(dbUser, nil).Once()
		repo.On("CreateSocial", ctx, mock.AnythingOfType("*user.UserSocial")).Return(nil).Once()

		res, err := svc.OAuthLogin(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, dbUser.Email, res.User.Email)

		repo.AssertExpectations(t)
	})

	t.Run("success create new user and social link", func(t *testing.T) {
		req := &OAuthLoginRequest{
			Provider:   "google",
			IDToken:    "mock-token",
			Email:      "new@example.com",
			FullName:   "New User",
			ProviderID: "google-uid-789",
		}

		repo.On("FindSocial", ctx, req.Provider, req.ProviderID).Return(nil, errors.New("not found")).Once()
		repo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("not found")).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil).Run(func(args mock.Arguments) {
			u := args.Get(1).(*User)
			u.ID = uuid.New()
		}).Once()
		repo.On("CreateSocial", ctx, mock.AnythingOfType("*user.UserSocial")).Return(nil).Once()

		res, err := svc.OAuthLogin(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Email, res.User.Email)

		repo.AssertExpectations(t)
	})
}

func TestService_UpdateAvatar(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		dbUser := &User{
			FullName: "Avatar User",
			Email:    "avatar@example.com",
		}
		dbUser.ID = userID

		mockAsset := &asset.Asset{
			FileURL: "http://localhost:9000/zerobin-avatars/avatars/"+userID.String()+".png",
		}
		mockAsset.ID = uuid.New()

		repo.On("FindByID", ctx, userID).Return(dbUser, nil).Once()
		assetSvc.On("Upload", ctx, userID.String()+".png", "avatars", mock.Anything, int64(10), "image/png", &userID).Return(mockAsset, nil).Once()
		repo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()

		res, err := svc.UpdateAvatar(ctx, userID, strings.NewReader("test-data"), 10, "image/png")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "http://localhost:9000/zerobin-avatars/avatars/"+userID.String()+".png", res.AvatarURL)

		repo.AssertExpectations(t)
		assetSvc.AssertExpectations(t)
	})
}

func TestService_RefreshToken(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		dbUser := &User{
			FullName: "Refresh User",
			Email:    "refresh@example.com",
			Role:     "user",
		}
		dbUser.ID = userID

		refreshToken, err := jwtSvc.GenerateRefreshToken(userID.String())
		assert.NoError(t, err)

		req := &RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		repo.On("FindByID", ctx, userID).Return(dbUser, nil).Once()

		res, err := svc.RefreshToken(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.Equal(t, dbUser.Email, res.User.Email)

		repo.AssertExpectations(t)
	})

	t.Run("invalid token error", func(t *testing.T) {
		req := &RefreshTokenRequest{
			RefreshToken: "invalid-token",
		}

		res, err := svc.RefreshToken(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "invalid or expired refresh token", err.Error())
	})
}

func TestService_ChangePassword(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()
	userID := uuid.New()

	hashedPassword, _ := auth.HashPassword("correctpassword")
	dbUser := &User{
		FullName: "Change Pass",
		Email:    "changepass@example.com",
		Password: string(hashedPassword),
	}
	dbUser.ID = userID

	t.Run("success", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "correctpassword",
			NewPassword:     "newsecurepassword",
		}
		repo.On("FindByID", ctx, userID).Return(dbUser, nil).Once()
		repo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()

		err := svc.ChangePassword(ctx, userID, req)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("invalid current password", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newsecurepassword",
		}
		repo.On("FindByID", ctx, userID).Return(dbUser, nil).Once()

		err := svc.ChangePassword(ctx, userID, req)
		assert.Error(t, err)
		assert.Equal(t, "invalid current password", err.Error())

		repo.AssertExpectations(t)
	})
}

func TestService_ForgotPassword(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	dbUser := &User{
		FullName: "Forgot User",
		Email:    "forgot@example.com",
	}
	dbUser.ID = uuid.New()

	t.Run("success exists", func(t *testing.T) {
		req := &ForgotPasswordRequest{
			Email: "forgot@example.com",
		}
		repo.On("FindByEmail", ctx, req.Email).Return(dbUser, nil).Once()
		repo.On("DeleteResetToken", ctx, req.Email).Return(nil).Once()
		repo.On("CreateResetToken", ctx, mock.AnythingOfType("*user.PasswordReset")).Return(nil).Once()
		emailSvc.On("SendTemplate", req.Email, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		err := svc.ForgotPassword(ctx, req)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
		emailSvc.AssertExpectations(t)
	})

	t.Run("not found prevents user enumeration", func(t *testing.T) {
		req := &ForgotPasswordRequest{
			Email: "notfound@example.com",
		}
		repo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("record not found")).Once()

		err := svc.ForgotPassword(ctx, req)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
	})
}

func TestService_ResetPassword(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	resetToken := "some-token"
	resetEmail := "reset@example.com"
	dbUser := &User{
		FullName: "Reset User",
		Email:    resetEmail,
	}
	dbUser.ID = uuid.New()

	t.Run("success", func(t *testing.T) {
		req := &ResetPasswordRequest{
			Token:    resetToken,
			Password: "newpassword123",
		}
		reset := &PasswordReset{
			Email:     resetEmail,
			Token:     resetToken,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		repo.On("FindResetToken", ctx, resetToken).Return(reset, nil).Once()
		repo.On("FindByEmail", ctx, resetEmail).Return(dbUser, nil).Once()
		repo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()
		repo.On("DeleteResetToken", ctx, resetEmail).Return(nil).Once()

		err := svc.ResetPassword(ctx, req)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("expired token", func(t *testing.T) {
		req := &ResetPasswordRequest{
			Token:    resetToken,
			Password: "newpassword123",
		}
		reset := &PasswordReset{
			Email:     resetEmail,
			Token:     resetToken,
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		}

		repo.On("FindResetToken", ctx, resetToken).Return(reset, nil).Once()
		repo.On("DeleteResetToken", ctx, resetEmail).Return(nil).Once()

		err := svc.ResetPassword(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, "invalid or expired reset token", err.Error())

		repo.AssertExpectations(t)
	})
}

func TestService_VerifyEmail(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	token := "verify-token"
	email := "verify@example.com"
	dbUser := &User{
		FullName:   "Verify User",
		Email:      email,
		IsVerified: false,
	}
	dbUser.ID = uuid.New()

	t.Run("success", func(t *testing.T) {
		req := &VerifyEmailRequest{
			Token: token,
		}
		verify := &EmailVerification{
			Email:     email,
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		repo.On("FindVerificationToken", ctx, token).Return(verify, nil).Once()
		repo.On("FindByEmail", ctx, email).Return(dbUser, nil).Once()
		repo.On("Update", ctx, mock.AnythingOfType("*user.User")).Return(nil).Once()
		repo.On("DeleteVerificationToken", ctx, email).Return(nil).Once()

		err := svc.VerifyEmail(ctx, req)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("expired token", func(t *testing.T) {
		req := &VerifyEmailRequest{
			Token: token,
		}
		verify := &EmailVerification{
			Email:     email,
			Token:     token,
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		}

		repo.On("FindVerificationToken", ctx, token).Return(verify, nil).Once()
		repo.On("DeleteVerificationToken", ctx, email).Return(nil).Once()

		err := svc.VerifyEmail(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, "invalid or expired verification token", err.Error())

		repo.AssertExpectations(t)
	})
}

func TestService_ResendVerification(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	email := "resend@example.com"
	dbUser := &User{
		FullName:   "Resend User",
		Email:      email,
		IsVerified: false,
	}
	dbUser.ID = uuid.New()

	t.Run("success", func(t *testing.T) {
		req := &ResendVerificationRequest{
			Email: email,
		}
		repo.On("FindByEmail", ctx, email).Return(dbUser, nil).Once()
		repo.On("DeleteVerificationToken", ctx, email).Return(nil).Once()
		repo.On("CreateVerificationToken", ctx, mock.AnythingOfType("*user.EmailVerification")).Return(nil).Once()
		emailSvc.On("SendTemplate", email, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()

		err := svc.ResendVerification(ctx, req)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
		emailSvc.AssertExpectations(t)
	})

	t.Run("already verified error", func(t *testing.T) {
		req := &ResendVerificationRequest{
			Email: email,
		}
		verifiedUser := &User{
			FullName:   "Verified User",
			Email:      email,
			IsVerified: true,
		}
		repo.On("FindByEmail", ctx, email).Return(verifiedUser, nil).Once()

		err := svc.ResendVerification(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, "email address is already verified", err.Error())

		repo.AssertExpectations(t)
	})
}

func TestService_Logout(t *testing.T) {
	repo := new(MockUserRepository)
	auditSvc := new(MockAuditService)
	jwtSvc := auth.NewJWTService(&configs.JWTConfig{
		Secret:        "test_secret_key_12345_67890_12345",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	storageSvc := new(MockStorage)
	emailSvc := new(MockEmailService)
	txManager := new(MockTransactionManager)
	assetSvc := new(MockAssetService)
	svc := NewService(repo, auditSvc, jwtSvc, storageSvc, emailSvc, txManager, assetSvc)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		token, err := jwtSvc.GenerateAccessToken(uuid.New().String(), "user")
		assert.NoError(t, err)

		repo.On("AddToBlacklist", ctx, mock.AnythingOfType("*user.JWTBlacklist")).Return(nil).Once()

		err = svc.Logout(ctx, token)
		assert.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("invalid token error", func(t *testing.T) {
		err := svc.Logout(ctx, "invalid-token-string")
		assert.Error(t, err)
	})
}

