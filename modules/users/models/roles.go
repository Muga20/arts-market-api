package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID         uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	RoleName   string    `gorm:"type:varchar(100);not null" json:"role_name"`
	RoleNumber int       `gorm:"type:int;not null" json:"role_number"`
	IsActive   bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set
func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
