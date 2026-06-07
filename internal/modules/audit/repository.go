package audit

import (
	"gorm.io/gorm"
)

type Repository interface {
	Create(audit *AuditTrail) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(audit *AuditTrail) error {
	return r.db.Create(audit).Error
}
