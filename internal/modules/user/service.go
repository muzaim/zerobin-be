package user

import (
	"context"
	"errors"
	"fmt"

	"cellar-backend/internal/modules/audit"
	"cellar-backend/pkg/auth"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, req *CreateUserRequest, operatorID string) (*UserResponse, error)
	GetAll(ctx context.Context) ([]UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest, operatorID string) (*UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID, operatorID string) error
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}

type service struct {
	repo       Repository
	auditServ  audit.Service
	jwtService *auth.JWTService
}

func NewService(repo Repository, auditServ audit.Service, jwtService *auth.JWTService) Service {
	return &service{
		repo:       repo,
		auditServ:  auditServ,
		jwtService: jwtService,
	}
}

func (s *service) Create(ctx context.Context, req *CreateUserRequest, operatorID string) (*UserResponse, error) {
	// Check if user already exists
	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email address is already in use")
	}

	// Encrypt password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	u := &User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	// Log audit trail
	description := fmt.Sprintf("Created user %s with ID %s", u.Email, u.ID.String())
	_ = s.auditServ.Record(operatorID, "User", "Create", description)

	res := ToUserResponse(u)
	return &res, nil
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

	// Log audit trail
	description := fmt.Sprintf("Updated user %s attributes", u.Email)
	_ = s.auditServ.Record(operatorID, "User", "Update", description)

	res := ToUserResponse(u)
	return &res, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, operatorID string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Log audit trail
	description := fmt.Sprintf("Deleted user %s with ID %s", u.Email, id.String())
	_ = s.auditServ.Record(operatorID, "User", "Delete", description)

	return nil
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Sign JWT access & refresh tokens
	accessToken, err := s.jwtService.GenerateAccessToken(u.ID.String(), u.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(u.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log audit trail
	description := fmt.Sprintf("User logged in successfully")
	_ = s.auditServ.Record(u.ID.String(), "Auth", "Login", description)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         ToUserResponse(u),
	}, nil
}
