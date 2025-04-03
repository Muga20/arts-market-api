package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Technique struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook (optional) to generate UUID or custom ID if needed
func (ai *Technique) BeforeCreate(tx *gorm.DB) (err error) {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	return
}
