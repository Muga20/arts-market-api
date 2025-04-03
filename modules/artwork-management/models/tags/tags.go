package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	TagName  string    `gorm:"type:varchar(255);unique;not null;column:tag_name" json:"tag_name"`
	IsActive bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
