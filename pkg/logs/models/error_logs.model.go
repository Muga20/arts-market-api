package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ErrorLog struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())"`
	Level          string    `gorm:"type:varchar(255);not null"`
	Message        string    `gorm:"type:varchar(255);not null"`
	StackTrace     string    `gorm:"type:text"`
	FileName       string    `gorm:"type:varchar(255)"`
	MethodName     string    `gorm:"type:varchar(255)"`
	LineNumber     int       `gorm:"type:int"`
	Metadata       JSON      `gorm:"type:json"`
	IPAddress      string    `gorm:"type:varchar(255)"`
	UserAgent      string    `gorm:"type:varchar(255)"`
	OccurredAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Attempts       int       `gorm:"type:int;default:1"`
	LastOccurredAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

// BeforeCreate hook
func (e *ErrorLog) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return
}

// JSON type for MySQL compatibility
type JSON map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, j)
}
