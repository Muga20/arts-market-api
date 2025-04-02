package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSecurity struct {
	ID                     uuid.UUID  `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID                 uuid.UUID  `gorm:"type:char(36);not null;index" json:"user_id"`
	Password               string     `gorm:"type:varchar(255);not null" json:"-"`
	IsEmailVerifiedAt      *time.Time `gorm:"type:timestamp" json:"is_email_verified_at,omitempty"`
	LoginAttempts          int        `gorm:"type:int;not null;default:0" json:"login_attempts"`
	LastLoginAttemptAt     *time.Time `gorm:"type:timestamp" json:"last_login_attempt_at,omitempty"`
	IsLocked               bool       `gorm:"type:boolean;not null;default:false" json:"is_locked"`
	LockedAt               *time.Time `gorm:"type:timestamp" json:"locked_at,omitempty"`
	UnlockToken            *string    `gorm:"type:varchar(255)" json:"unlock_token,omitempty"`
	UnlockTokenExpiresAt   *time.Time `gorm:"type:timestamp" json:"unlock_token_expires_at,omitempty"`
	LastSuccessfulLoginAt  *time.Time `gorm:"type:timestamp" json:"last_successful_login_at,omitempty"`
	PasswordResetToken     *string    `gorm:"type:varchar(255)" json:"-"`
	PasswordResetExpiresAt *time.Time `gorm:"type:timestamp" json:"password_reset_expires_at,omitempty"`
	CreatedAt              time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updated_at"`

	// Foreign key relation
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (u *UserSecurity) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
