package user

import (
	"context"

	"cellar-backend/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	FindAll(ctx context.Context) ([]User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindSocial(ctx context.Context, provider string, providerID string) (*UserSocial, error)
	CreateSocial(ctx context.Context, social *UserSocial) error
	CreateResetToken(ctx context.Context, reset *PasswordReset) error
	FindResetToken(ctx context.Context, token string) (*PasswordReset, error)
	DeleteResetToken(ctx context.Context, email string) error
	CreateVerificationToken(ctx context.Context, verify *EmailVerification) error
	FindVerificationToken(ctx context.Context, token string) (*EmailVerification, error)
	DeleteVerificationToken(ctx context.Context, email string) error
	AddToBlacklist(ctx context.Context, blacklist *JWTBlacklist) error
	IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, u *User) error {
	return database.GetDB(ctx, r.db).Create(u).Error
}

func (r *repository) FindAll(ctx context.Context) ([]User, error) {
	var users []User
	err := database.GetDB(ctx, r.db).Find(&users).Error
	return users, err
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := database.GetDB(ctx, r.db).First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := database.GetDB(ctx, r.db).First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) Update(ctx context.Context, u *User) error {
	return database.GetDB(ctx, r.db).Save(u).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return database.GetDB(ctx, r.db).Delete(&User{}, "id = ?", id).Error
}

func (r *repository) FindSocial(ctx context.Context, provider string, providerID string) (*UserSocial, error) {
	var s UserSocial
	err := database.GetDB(ctx, r.db).First(&s, "provider = ? AND provider_id = ?", provider, providerID).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) CreateSocial(ctx context.Context, social *UserSocial) error {
	return database.GetDB(ctx, r.db).Create(social).Error
}

func (r *repository) CreateResetToken(ctx context.Context, reset *PasswordReset) error {
	return database.GetDB(ctx, r.db).Create(reset).Error
}

func (r *repository) FindResetToken(ctx context.Context, token string) (*PasswordReset, error) {
	var pr PasswordReset
	err := database.GetDB(ctx, r.db).First(&pr, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *repository) DeleteResetToken(ctx context.Context, email string) error {
	return database.GetDB(ctx, r.db).Delete(&PasswordReset{}, "email = ?", email).Error
}

func (r *repository) CreateVerificationToken(ctx context.Context, verify *EmailVerification) error {
	return database.GetDB(ctx, r.db).Create(verify).Error
}

func (r *repository) FindVerificationToken(ctx context.Context, token string) (*EmailVerification, error) {
	var ev EmailVerification
	err := database.GetDB(ctx, r.db).First(&ev, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (r *repository) DeleteVerificationToken(ctx context.Context, email string) error {
	return database.GetDB(ctx, r.db).Delete(&EmailVerification{}, "email = ?", email).Error
}

func (r *repository) AddToBlacklist(ctx context.Context, blacklist *JWTBlacklist) error {
	return database.GetDB(ctx, r.db).Create(blacklist).Error
}

func (r *repository) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	err := database.GetDB(ctx, r.db).Model(&JWTBlacklist{}).Where("token_hash = ?", tokenHash).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

