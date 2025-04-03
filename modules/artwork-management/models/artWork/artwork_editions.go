package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Edition struct {
	ID            uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID     uuid.UUID `gorm:"type:char(36);not null" json:"artwork_id"`
	EditionNumber int       `gorm:"type:int;not null" json:"edition_number"`
	TotalEditions int       `gorm:"type:int;not null" json:"total_editions"`
	Status        string    `gorm:"type:enum('available','sold','reserved');default:'available'" json:"status"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Artwork Artwork `gorm:"foreignKey:ArtworkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (e *Edition) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
