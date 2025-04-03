package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Attribute struct {
	ID            uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	AttributeName string    `gorm:"type:varchar(255);not null" json:"attribute_name"`
	Type          string    `gorm:"type:enum('string','integer','float','boolean');default:'string'" json:"type"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (a *Attribute) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
