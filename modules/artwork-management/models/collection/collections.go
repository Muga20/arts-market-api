package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/muga20/artsMarket/modules/users/models"
	"gorm.io/gorm"
)

type CollectionStatus string

const (
	DraftStatus     CollectionStatus = "draft"
	PublishedStatus CollectionStatus = "published"
	ArchivedStatus  CollectionStatus = "archived"
)

type Collection struct {
	ID          uuid.UUID        `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID      uuid.UUID        `gorm:"type:char(36);not null;index" json:"user_id"`
	Name        string           `gorm:"type:varchar(255);not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Status      CollectionStatus `gorm:"type:enum('draft', 'published', 'archived');default:'draft'" json:"status"`

	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`

	CoverImageURL   string `gorm:"type:varchar(255)" json:"cover_image_url"`
	PrimaryImageURL string `gorm:"type:varchar(255)" json:"primary_image_url"`

	User models.User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type CreateCollectionRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type CollectionResponse struct {
	ID            uuid.UUID        `json:"id"`
	UserID        uuid.UUID        `json:"user_id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Status        CollectionStatus `json:"status"`
	CoverImageURL string           `json:"cover_image_url"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// BeforeCreate hook to generate default values if not set
func (ai *Collection) BeforeCreate(tx *gorm.DB) (err error) {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	return
}
