package asset

import (
	"zerobin-be/internal/entities"

	"github.com/google/uuid"
)

type Asset struct {
	entities.Base
	FileName     string     `gorm:"type:varchar(255);not null" json:"file_name"`
	OriginalName string     `gorm:"type:varchar(255);not null" json:"original_name"`
	FilePath     string     `gorm:"type:varchar(255);not null" json:"file_path"`
	FileURL      string     `gorm:"type:varchar(512);not null" json:"file_url"`
	FileSize     int64      `gorm:"not null" json:"file_size"`
	ContentType  string     `gorm:"type:varchar(100);not null" json:"content_type"`
	UploadedByID *uuid.UUID `gorm:"type:uuid" json:"uploaded_by_id"`
}

func (Asset) TableName() string {
	return "sys_assets"
}
