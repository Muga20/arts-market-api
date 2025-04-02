package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSession struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID       uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	SessionToken string    `gorm:"type:varchar(255);not null" json:"session_token"`
	DeviceInfo   string    `gorm:"type:varchar(255)" json:"device_info"` // e.g., "iPhone 14, iOS 16"
	IPAddress    string    `gorm:"type:varchar(45)" json:"ip_address"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	ExpiresAt    time.Time `gorm:"type:timestamp;not null" json:"expires_at"`

	// Foreign key relation
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (s *UserSession) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
