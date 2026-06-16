package blog

import (
	"zerobin-be/internal/entities"
	"zerobin-be/internal/modules/user"

	"github.com/google/uuid"
)

type Blog struct {
	entities.Base
	UserID   *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Author   *user.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"author,omitempty"`
	Title    string     `gorm:"type:varchar(200);not null" json:"title"`
	Slug     string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Desc     string     `gorm:"type:text;not null" json:"desc"`
	Category string     `gorm:"type:varchar(50);index;not null" json:"category"`
}

func (Blog) TableName() string {
	return "mst_blogs"
}
