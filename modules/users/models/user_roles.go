package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID   uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	RoleID   uuid.UUID `gorm:"type:char(36);not null;index" json:"role_id"`
	IsActive bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`

	// Foreign key relations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Role Role `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (ur *UserRole) BeforeCreate(tx *gorm.DB) (err error) {
	if ur.ID == uuid.Nil {
		ur.ID = uuid.New()
	}
	return
}
