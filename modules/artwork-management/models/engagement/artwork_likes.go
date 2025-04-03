package models

import (
	"github.com/google/uuid"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	user "github.com/muga20/artsMarket/modules/users/models"
	"gorm.io/gorm"
	"time"
)

type ArtworkLike struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID uuid.UUID `gorm:"type:char(36);not null;index" json:"artwork_id"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	LikedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"liked_at"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Artwork art.Artwork `gorm:"foreignKey:ArtworkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User    user.User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Unique constraint to prevent duplicate likes
func (ArtworkLike) TableUnique() []string {
	return []string{"artwork_id", "user_id"}
}

// BeforeCreate hook to generate default values if not set
func (ai *ArtworkLike) BeforeSave(tx *gorm.DB) (err error) {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	return
}
