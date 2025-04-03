package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArtworkImage model to handle multiple images
type ArtworkImage struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID uuid.UUID `gorm:"type:char(36);not null" json:"artwork_id"`
	ImageURL  string    `gorm:"type:varchar(255);not null" json:"image_url"`
	IsPrimary bool      `gorm:"type:boolean;not null;default:false" json:"is_primary"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (ai *ArtworkImage) BeforeCreate(tx *gorm.DB) (err error) {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	return
}
