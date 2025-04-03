package models

import (
	"github.com/google/uuid"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	user "github.com/muga20/artsMarket/modules/users/models"
	"gorm.io/gorm"
	"time"
)

type ArtworkView struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID uuid.UUID `gorm:"type:char(36);not null;index" json:"artwork_id"`
	UserID    uuid.UUID `gorm:"type:char(36);index" json:"user_id"`

	// Additional tracking fields
	IPAddress   string `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent   string `gorm:"type:text" json:"user_agent"`
	ReferrerURL string `gorm:"type:text" json:"referrer_url"`
	CountryCode string `gorm:"type:varchar(2)" json:"country_code"`
	DeviceType  string `gorm:"type:varchar(20)" json:"device_type"`

	ViewedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index" json:"viewed_at"`

	// Timestamps
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Artwork art.Artwork `gorm:"foreignKey:ArtworkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User    user.User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// BeforeCreate hook
func (v *ArtworkView) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return
}
