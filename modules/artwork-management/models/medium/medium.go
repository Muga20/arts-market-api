package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Medium struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	MediumName  string    `gorm:"type:varchar(255);not null" json:"medium_name"`
	Description string    `gorm:"type:text" json:"description"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate default values if not set
func (ai *Medium) BeforeCreate(tx *gorm.DB) (err error) {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	return
}
