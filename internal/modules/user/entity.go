package user

import (
	"time"

	"cellar-backend/internal/entities"

	"github.com/google/uuid"
)

type User struct {
	entities.Base
	FullName   string `gorm:"type:varchar(100);not null" json:"full_name"`
	Email      string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password   string `gorm:"type:varchar(255)" json:"-"`
	Role       string `gorm:"type:varchar(50);default:'user';index;not null" json:"role"`
	AvatarURL  string `gorm:"type:varchar(255)" json:"avatar_url"`
	IsVerified bool   `gorm:"default:false;not null" json:"is_verified"`
}

func (User) TableName() string {
	return "sys_users"
}

type UserSocial struct {
	entities.Base
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider   string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_provider_id" json:"provider"`
	ProviderID string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_id" json:"provider_id"`
}

func (UserSocial) TableName() string {
	return "sys_user_socials"
}

type PasswordReset struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(100);index;not null" json:"email"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (PasswordReset) TableName() string {
	return "sys_password_resets"
}

type EmailVerification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (EmailVerification) TableName() string {
	return "sys_email_verifications"
}

type JWTBlacklist struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TokenHash string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"token_hash"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (JWTBlacklist) TableName() string {
	return "sys_jwt_blacklist"
}

