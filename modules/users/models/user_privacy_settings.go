package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserPrivacySetting defines the user's privacy settings
type UserPrivacySetting struct {
	ID                  uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID              uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	ShowEmail           bool      `gorm:"type:boolean;not null;default:false" json:"show_email"`
	ShowPhone           bool      `gorm:"type:boolean;not null;default:false" json:"show_phone"`
	ShowLocation        bool      `gorm:"type:boolean;not null;default:false" json:"show_location"`
	AllowFollowRequests bool      `gorm:"type:boolean;not null;default:true" json:"allow_follow_requests"`
	AllowMessagesFrom   string    `gorm:"type:varchar(20);not null;default:'everyone'" json:"allow_messages_from"`

	// Foreign key relation
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (u *UserPrivacySetting) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

// CreateDefaultPrivacySetting ensures a user always has a privacy setting entry
func CreateDefaultPrivacySetting(db *gorm.DB, userID uuid.UUID) error {
	var setting UserPrivacySetting
	err := db.Where("user_id = ?", userID).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		// Create entry with defaults
		setting = UserPrivacySetting{
			ID:                  uuid.New(),
			UserID:              userID,
			ShowEmail:           false,
			ShowPhone:           false,
			ShowLocation:        false,
			AllowFollowRequests: true,
			AllowMessagesFrom:   "everyone",
		}
		return db.Create(&setting).Error
	}
	return err
}
