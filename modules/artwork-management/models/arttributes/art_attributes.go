package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArtworkAttribute struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID   uuid.UUID `gorm:"type:char(36);not null;index" json:"artwork_id"`
	AttributeID uuid.UUID `gorm:"type:char(36);not null;index" json:"attribute_id"`
	Value       string    `gorm:"type:varchar(255)" json:"value"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	// Foreign key relationships
	//Artwork   Artwork   `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE;" json:"-"`
	Attribute Attribute `gorm:"foreignKey:AttributeID;constraint:OnDelete:CASCADE;" json:"-"`
}

// BeforeCreate hook to generate UUID if not set
func (a *ArtworkAttribute) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
