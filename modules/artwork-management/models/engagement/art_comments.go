package models

import (
	"github.com/google/uuid"
	art "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	user "github.com/muga20/artsMarket/modules/users/models"
	"gorm.io/gorm"
	"time"
)

type ArtworkComment struct {
	ID        uuid.UUID  `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	ArtworkID uuid.UUID  `gorm:"type:char(36);not null;index" json:"artwork_id"`
	UserID    uuid.UUID  `gorm:"type:char(36);not null;index" json:"user_id"`
	ParentID  *uuid.UUID `gorm:"type:char(36);index" json:"parent_id"` // For nested comments
	Content   string     `gorm:"type:text;not null" json:"content"`
	IsEdited  bool       `gorm:"type:boolean;default:false" json:"is_edited"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Artwork art.Artwork      `gorm:"foreignKey:ArtworkID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User    user.User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Parent  *ArtworkComment  `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Replies []ArtworkComment `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ai *ArtworkLike) BeforeCreate(tx *gorm.DB) (err error) {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	return
}
