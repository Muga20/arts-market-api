package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDetail struct {
	ID                uuid.UUID  `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID            uuid.UUID  `gorm:"type:char(36);not null;index" json:"user_id"`
	FirstName         string     `gorm:"type:varchar(100);not null" json:"first_name"`
	MiddleName        *string    `gorm:"type:varchar(100)" json:"middle_name,omitempty"`
	LastName          string     `gorm:"type:varchar(100);not null" json:"last_name"`
	Gender            *string    `gorm:"type:varchar(50)" json:"gender,omitempty"`
	DateOfBirth       *time.Time `gorm:"type:date" json:"date_of_birth,omitempty"`
	ProfileImage      string     `gorm:"type:text;not null" json:"profile_image"`
	CoverImage        string     `gorm:"type:text;not null" json:"cover_image"`
	AboutTheUser      *string    `gorm:"type:text" json:"about_the_user,omitempty"`
	IsProfilePublic   bool       `gorm:"type:boolean;not null;default:false" json:"is_profile_public"`
	Nickname          *string    `gorm:"type:varchar(100)" json:"nickname,omitempty"`
	PreferredPronouns *string    `gorm:"type:varchar(50)" json:"preferred_pronouns,omitempty"`
	CreatedAt         time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`

	// Foreign key relation
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (u *UserDetail) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
