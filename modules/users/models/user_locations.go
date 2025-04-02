package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserLocation struct {
	ID        uuid.UUID  `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID    *uuid.UUID `gorm:"type:char(36);index" json:"user_id,omitempty"` // UserID can be null, hence pointer
	Country   string     `gorm:"type:varchar(100)" json:"country"`
	State     string     `gorm:"type:varchar(100)" json:"state"`
	StateName string     `gorm:"type:varchar(100)" json:"state_name"`
	Continent string     `gorm:"type:varchar(100)" json:"continent"`
	City      string     `gorm:"type:varchar(100)" json:"city"`
	Zip       string     `gorm:"type:varchar(20)" json:"zip"`

	// Foreign key relation (allowing null for UserID)
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// BeforeCreate hook to generate UUID if not set
func (ul *UserLocation) BeforeCreate(tx *gorm.DB) (err error) {
	if ul.ID == uuid.Nil {
		ul.ID = uuid.New()
	}
	return
}
