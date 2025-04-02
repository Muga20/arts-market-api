package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Follower struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	FollowerID  uuid.UUID `gorm:"type:char(36);not null;index" json:"follower_id"`
	FollowingID uuid.UUID `gorm:"type:char(36);not null;index" json:"following_id"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Foreign key relations
	FollowerUser  User `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE"`
	FollowingUser User `gorm:"foreignKey:FollowingID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (f *Follower) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}
