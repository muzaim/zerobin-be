package blog

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, b *Blog) error
	FindAll(ctx context.Context, category string, authorID string, search string) ([]Blog, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Blog, error)
	FindBySlug(ctx context.Context, slug string) (*Blog, error)
	Update(ctx context.Context, b *Blog) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, b *Blog) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *repository) FindAll(ctx context.Context, category string, authorID string, search string) ([]Blog, error) {
	var blogs []Blog
	query := r.db.WithContext(ctx).Preload("Author")

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if authorID != "" {
		query = query.Where("user_id = ?", authorID)
	}
	if search != "" {
		dialect := r.db.Dialector.Name()
		if dialect == "sqlite" {
			query = query.Where("title LIKE ? OR slug LIKE ?", "%"+search+"%", "%"+search+"%")
		} else {
			query = query.Where("title ILIKE ? OR slug ILIKE ?", "%"+search+"%", "%"+search+"%")
		}
	}

	err := query.Find(&blogs).Error
	return blogs, err
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Blog, error) {
	var b Blog
	err := r.db.WithContext(ctx).Preload("Author").First(&b, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*Blog, error) {
	var b Blog
	err := r.db.WithContext(ctx).Preload("Author").First(&b, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) Update(ctx context.Context, b *Blog) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Blog{}, "id = ?", id).Error
}
