package user

import (
	"cellar-backend/internal/entities"
)

type User struct {
	entities.Base
	FullName string `gorm:"type:varchar(100);not null" json:"full_name"`
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	Role     string `gorm:"type:varchar(50);default:'user';not null" json:"role"`
}
