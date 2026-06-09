package user

import (
	"time"
)

type CreateUserRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email" validate:"omitempty,email"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID         string    `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ToUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:         u.ID.String(),
		FullName:   u.FullName,
		Email:      u.Email,
		Role:       u.Role,
		AvatarURL:  u.AvatarURL,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func ToUserResponses(users []User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = ToUserResponse(&u)
	}
	return res
}

type OAuthLoginRequest struct {
	Provider   string `json:"provider" validate:"required,oneof=google apple"`
	IDToken    string `json:"id_token" validate:"required"`
	Email      string `json:"email" validate:"omitempty,email"`
	FullName   string `json:"full_name"`
	ProviderID string `json:"provider_id"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PolicyRequest struct {
	PType string   `json:"ptype" validate:"required,oneof=p g"`
	Rule  []string `json:"rule" validate:"required,min=2"`
}

type PolicyRuleResponse struct {
	PType string   `json:"ptype"`
	Rule  []string `json:"rule"`
}
