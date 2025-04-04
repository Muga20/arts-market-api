package models

import (
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtworkTag struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID uuid.UUID `gorm:"type:char(36);not null;index" json:"artwork_id"`
	TagID     uuid.UUID `gorm:"type:char(36);not null;index" json:"tag_id"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	// Foreign key relations
	Artwork art.Artwork `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE"`
	Tag     Tag         `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (a *ArtworkTag) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
