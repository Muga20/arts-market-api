package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlockedUser struct {
	ID            uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID        uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`         // Who blocked
	BlockedUserID uuid.UUID `gorm:"type:char(36);not null;index" json:"blocked_user_id"` // Who is blocked
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Foreign key relations
	User        User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	BlockedUser User `gorm:"foreignKey:BlockedUserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (b *BlockedUser) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
