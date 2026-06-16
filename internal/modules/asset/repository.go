package asset

import (
	"context"

	"zerobin-be/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, a *Asset) error
	FindByID(ctx context.Context, id uuid.UUID) (*Asset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, a *Asset) error {
	return database.GetDB(ctx, r.db).Create(a).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Asset, error) {
	var a Asset
	err := database.GetDB(ctx, r.db).First(&a, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return database.GetDB(ctx, r.db).Delete(&Asset{}, "id = ?", id).Error
}
