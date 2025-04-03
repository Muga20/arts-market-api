package models

import (
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtworkCategory struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID  uuid.UUID `gorm:"type:char(36);not null;index" json:"artwork_id"`
	CategoryID uuid.UUID `gorm:"type:char(36);not null;index" json:"category_id"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	// Foreign key relations
	Artwork  art.Artwork `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE" json:"artwork"`
	Category Category    `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"category"`
}

// BeforeCreate hook to generate UUID if not set
func (ac *ArtworkCategory) BeforeCreate(tx *gorm.DB) (err error) {
	if ac.ID == uuid.Nil {
		ac.ID = uuid.New()
	}
	return
}
