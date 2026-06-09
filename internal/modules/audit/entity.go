package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditUser struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (AuditUser) TableName() string {
	return "sys_users"
}

type AuditTrail struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       *uuid.UUID  `gorm:"type:uuid;index" json:"user_id"`
	User         *AuditUser  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"user,omitempty"`
	Module       string      `gorm:"type:varchar(100);index" json:"module"`
	Action       string      `gorm:"type:varchar(100);index" json:"action"`
	ResourceID   string      `gorm:"type:varchar(255)" json:"resource_id,omitempty"`
	Description  string      `gorm:"type:text" json:"description"`
	IPAddress    string      `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent    string      `gorm:"type:text" json:"user_agent,omitempty"`
	Method       string      `gorm:"type:varchar(10)" json:"method,omitempty"`
	Path         string      `gorm:"type:text" json:"path,omitempty"`
	RequestID    string      `gorm:"type:varchar(100);index" json:"request_id,omitempty"`
	Payload      string      `gorm:"type:text" json:"payload,omitempty"`
	Status        string      `gorm:"type:varchar(20);index" json:"status"`
	StatusCode    int         `gorm:"type:integer" json:"status_code,omitempty"`
	ExecutionTime int64       `gorm:"type:bigint" json:"execution_time_ms,omitempty"`
	ErrorMessage  string      `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

func (a *AuditTrail) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (AuditTrail) TableName() string {
	return "sys_audit_trails"
}
