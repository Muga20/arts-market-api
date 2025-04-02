package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	Email       string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	PhoneNumber string         `gorm:"type:varchar(50)" json:"phone_number"`
	Username    string         `gorm:"type:varchar(50)" json:"username"`
	Status      string         `gorm:"type:varchar(50)" json:"status"`
	AuthType    string         `gorm:"type:varchar(50)" json:"auth_type"`
	IsActive    bool           `gorm:"type:boolean;not null;default:true" json:"is_active"`
	DeletedAt   gorm.DeletedAt `gorm:"type:timestamp;index" json:"deleted_at,omitempty"`
	CreatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
