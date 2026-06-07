package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditTrail struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(255)" json:"user_id"`
	Module      string    `gorm:"type:varchar(100)" json:"module"`
	Action      string    `gorm:"type:varchar(100)" json:"action"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (a *AuditTrail) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
