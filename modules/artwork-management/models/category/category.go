package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	CategoryName string    `gorm:"type:varchar(255);not null" json:"category_name"`
	Description  string    `gorm:"type:text" json:"description"`
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
