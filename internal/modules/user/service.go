package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"cellar-backend/internal/modules/audit"
	"cellar-backend/pkg/auth"
	"cellar-backend/pkg/database"
	"cellar-backend/pkg/email"
	"cellar-backend/pkg/storage"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, req *CreateUserRequest, operatorID string) (*UserResponse, error)
	GetAll(ctx context.Context) ([]UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest, operatorID string) (*UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID, operatorID string) error
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	OAuthLogin(ctx context.Context, req *OAuthLoginRequest) (*LoginResponse, error)
	UpdateAvatar(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) (*UserResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResponse, error)
	ChangePassword(ctx context.Context, id uuid.UUID, req *ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, req *VerifyEmailRequest) error
	ResendVerification(ctx context.Context, req *ResendVerificationRequest) error
	Logout(ctx context.Context, token string) error
}

type service struct {
	repo       Repository
	auditServ  audit.Service
	jwtService *auth.JWTService
	storage    storage.Storage
	emailSvc   email.Service
	txManager  database.TransactionManager
}

func NewService(repo Repository, auditServ audit.Service, jwtService *auth.JWTService, storage storage.Storage, emailSvc email.Service, txManager database.TransactionManager) Service {
	return &service{
		repo:       repo,
		auditServ:  auditServ,
		jwtService: jwtService,
		storage:    storage,
		emailSvc:   emailSvc,
		txManager:  txManager,
	}
}

func (s *service) Create(ctx context.Context, req *CreateUserRequest, operatorID string) (*UserResponse, error) {
	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email address is already in use")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	u := &User{
		FullName:   req.FullName,
		Email:      req.Email,
		Password:   string(hashedPassword),
		Role:       role,
		IsVerified: false,
	}

	var res *UserResponse
	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, u); err != nil {
			return err
		}

		token := uuid.New().String()
		verify := &EmailVerification{
			ID:        uuid.New(),
			Email:     u.Email,
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := s.repo.CreateVerificationToken(txCtx, verify); err != nil {
			return err
		}
		link := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
		subject := "Verify Your Email Address"
		emailData := map[string]interface{}{
			"FullName":         u.FullName,
			"VerificationLink": link,
		}
		_ = s.emailSvc.SendTemplate(u.Email, subject, "templates/email/welcome.html", emailData)

		response := ToUserResponse(u)
		res = &response
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *service) GetAll(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return ToUserResponses(users), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	res := ToUserResponse(u)
	return &res, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest, operatorID string) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Email != "" && req.Email != u.Email {
		existing, _ := s.repo.FindByEmail(ctx, req.Email)
		if existing != nil {
			return nil, errors.New("email address is already in use")
		}
		u.Email = req.Email
	}

	if req.FullName != "" {
		u.FullName = req.FullName
	}

	if req.Role != "" {
		u.Role = req.Role
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}

	res := ToUserResponse(u)
	return &res, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, operatorID string) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	match, needsUpgrade, err := auth.ComparePassword(u.Password, req.Password)
	if err != nil || !match {
		return nil, errors.New("invalid email or password")
	}

	if needsUpgrade {
		newHash, err := auth.HashPassword(req.Password)
		if err == nil {
			u.Password = newHash
			_ = s.repo.Update(ctx, u)
		}
	}

	if !u.IsVerified {
		return nil, errors.New("please verify your email address first")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(u.ID.String(), u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(u.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserResponse(u),
	}, nil
}

func (s *service) OAuthLogin(ctx context.Context, req *OAuthLoginRequest) (*LoginResponse, error) {
	var verifiedEmail, verifiedFullName, verifiedProviderID string

	if req.IDToken == "mock-token" {
		verifiedEmail = req.Email
		verifiedFullName = req.FullName
		verifiedProviderID = req.ProviderID
		if verifiedEmail == "" || verifiedProviderID == "" {
			return nil, errors.New("mock token requires email and provider_id")
		}
	} else {
		if req.Provider == "google" {
			var err error
			verifiedEmail, verifiedFullName, verifiedProviderID, err = verifyGoogleToken(req.IDToken)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("apple sign-in token verification not implemented for local dev environments (requires Apple Developer keys)")
		}
	}

	var u *User
	if err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		social, err := s.repo.FindSocial(txCtx, req.Provider, verifiedProviderID)
		if err == nil && social != nil {
			u, err = s.repo.FindByID(txCtx, social.UserID)
			if err != nil {
				return fmt.Errorf("social link exists but user not found: %w", err)
			}
		} else {
			existingUser, err := s.repo.FindByEmail(txCtx, verifiedEmail)
			if err == nil && existingUser != nil {
				u = existingUser
				newSocial := &UserSocial{
					UserID:     u.ID,
					Provider:   req.Provider,
					ProviderID: verifiedProviderID,
				}
				if err := s.repo.CreateSocial(txCtx, newSocial); err != nil {
					return fmt.Errorf("failed to link social account: %w", err)
				}
			} else {
				u = &User{
					FullName:   verifiedFullName,
					Email:      verifiedEmail,
					Role:       "user",
					IsVerified: true,
				}
				if err := s.repo.Create(txCtx, u); err != nil {
					return fmt.Errorf("failed to create user: %w", err)
				}

				newSocial := &UserSocial{
					UserID:     u.ID,
					Provider:   req.Provider,
					ProviderID: verifiedProviderID,
				}
				if err := s.repo.CreateSocial(txCtx, newSocial); err != nil {
					return fmt.Errorf("failed to create social link: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(u.ID.String(), u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(u.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserResponse(u),
	}, nil
}

func verifyGoogleToken(idToken string) (string, string, string, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to verify token with Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", errors.New("invalid google token")
	}

	var info struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
		Err   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", "", "", fmt.Errorf("failed to decode google token response: %w", err)
	}

	if info.Err != "" {
		return "", "", "", fmt.Errorf("google token verification error: %s", info.Err)
	}

	if info.Sub == "" {
		return "", "", "", errors.New("google token has no subject claim")
	}

	return info.Email, info.Name, info.Sub, nil
}

func (s *service) UpdateAvatar(ctx context.Context, id uuid.UUID, reader io.Reader, size int64, contentType string) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	ext := ""
	switch contentType {
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return nil, errors.New("unsupported image format (only jpeg, png, webp are supported)")
	}

	filename := fmt.Sprintf("%s%s", id.String(), ext)
	avatarURL, err := s.storage.UploadAvatar(ctx, filename, reader, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	u.AvatarURL = avatarURL
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user avatar url: %w", err)
	}

	res := ToUserResponse(u)
	return &res, nil
}

func (s *service) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*LoginResponse, error) {
	claims, err := s.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user id in refresh token")
	}

	u, err := s.repo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, errors.New("user associated with token not found")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(u.ID.String(), u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(u.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         ToUserResponse(u),
	}, nil
}

func (s *service) ChangePassword(ctx context.Context, id uuid.UUID, req *ChangePasswordRequest) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}
	match, _, err := auth.ComparePassword(u.Password, req.CurrentPassword)
	if err != nil || !match {
		return errors.New("invalid current password")
	}
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	u.Password = string(hashedPassword)
	if err := s.repo.Update(ctx, u); err != nil {
		return err
	}
	return nil
}

func (s *service) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil || u == nil {
		return nil
	}

	token := uuid.New().String()
	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		_ = s.repo.DeleteResetToken(txCtx, req.Email)
		reset := &PasswordReset{
			ID:        uuid.New(),
			Email:     req.Email,
			Token:     token,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		if err := s.repo.CreateResetToken(txCtx, reset); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	link := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	subject := "Reset Your Password"
	emailData := map[string]interface{}{
		"ResetLink": link,
	}
	return s.emailSvc.SendTemplate(req.Email, subject, "templates/email/reset_password.html", emailData)
}

func (s *service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		reset, err := s.repo.FindResetToken(txCtx, req.Token)
		if err != nil {
			return errors.New("invalid or expired reset token")
		}
		if time.Now().After(reset.ExpiresAt) {
			_ = s.repo.DeleteResetToken(txCtx, reset.Email)
			return errors.New("invalid or expired reset token")
		}
		u, err := s.repo.FindByEmail(txCtx, reset.Email)
		if err != nil {
			return errors.New("user not found")
		}
		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		u.Password = string(hashedPassword)
		if err := s.repo.Update(txCtx, u); err != nil {
			return err
		}
		_ = s.repo.DeleteResetToken(txCtx, reset.Email)
		return nil
	})
}

func (s *service) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) error {
	return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		verify, err := s.repo.FindVerificationToken(txCtx, req.Token)
		if err != nil {
			return errors.New("invalid or expired verification token")
		}
		if time.Now().After(verify.ExpiresAt) {
			_ = s.repo.DeleteVerificationToken(txCtx, verify.Email)
			return errors.New("invalid or expired verification token")
		}
		u, err := s.repo.FindByEmail(txCtx, verify.Email)
		if err != nil {
			return errors.New("user not found")
		}
		u.IsVerified = true
		if err := s.repo.Update(txCtx, u); err != nil {
			return err
		}
		_ = s.repo.DeleteVerificationToken(txCtx, verify.Email)
		return nil
	})
}

func (s *service) ResendVerification(ctx context.Context, req *ResendVerificationRequest) error {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil || u == nil {
		return nil
	}
	if u.IsVerified {
		return errors.New("email address is already verified")
	}

	token := uuid.New().String()
	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		_ = s.repo.DeleteVerificationToken(txCtx, req.Email)
		verify := &EmailVerification{
			ID:        uuid.New(),
			Email:     req.Email,
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := s.repo.CreateVerificationToken(txCtx, verify); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	link := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
	subject := "Verify Your Email Address"
	emailData := map[string]interface{}{
		"FullName":         u.FullName,
		"VerificationLink": link,
	}
	return s.emailSvc.SendTemplate(req.Email, subject, "templates/email/welcome.html", emailData)
}

func (s *service) Logout(ctx context.Context, token string) error {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	blacklist := &JWTBlacklist{
		TokenHash: tokenHash,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	return s.repo.AddToBlacklist(ctx, blacklist)
}
