package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID                uuid.UUID       `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID            uuid.UUID       `gorm:"type:char(36);not null;index" json:"user_id"`
	SenderID          *uuid.UUID      `gorm:"type:char(36);index" json:"sender_id,omitempty"`
	NotificationType  string          `gorm:"type:varchar(50);not null" json:"notification_type"`
	Message           string          `gorm:"type:text;not null" json:"message"`
	IsRead            bool            `gorm:"type:boolean;not null;default:false" json:"is_read"`
	IsSystemGenerated bool            `gorm:"type:boolean;not null;default:false" json:"is_system_generated"`
	EntityType        string          `gorm:"type:varchar(50);index:idx_entity" json:"entity_type"`
	EntityID          *uuid.UUID      `gorm:"type:char(36);index:idx_entity" json:"entity_id,omitempty"`
	Priority          int             `gorm:"type:int;not null;default:0" json:"priority"`
	Metadata          json.RawMessage `gorm:"type:json" json:"metadata"`
	ExpiresAt         *time.Time      `gorm:"type:timestamp" json:"expires_at,omitempty"`
	CreatedAt         time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return
}

// ToJSON serializes the Notification to JSON format
func (n *Notification) ToJSON() ([]byte, error) {
	data, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return data, nil
}
