package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type Category struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	CategoryName string    `gorm:"type:varchar(255);not null" json:"category_name"`
	Slug         string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	Description  string    `gorm:"type:text" json:"description"`
	IsActive     bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// BeforeCreate hook to generate default values
func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	// Generate slug from category name if empty
	if c.Slug == "" {
		c.Slug = slug.Make(c.CategoryName)

		// Ensure slug is unique by appending UUID if needed
		var count int64
		tx.Model(&Category{}).
			Where("slug = ?", c.Slug).
			Count(&count)

		if count > 0 {
			c.Slug = slug.Make(c.CategoryName + "-" + strings.Split(c.ID.String(), "-")[0])
		}
	}
	return
}

// BeforeUpdate hook to update slug if category name changes
func (c *Category) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed("CategoryName") {
		newSlug := slug.Make(c.CategoryName)
		if newSlug != c.Slug {
			c.Slug = newSlug

			// Check for uniqueness
			var count int64
			tx.Model(&Category{}).
				Where("slug = ? AND id != ?", c.Slug, c.ID).
				Count(&count)

			if count > 0 {
				c.Slug = slug.Make(c.CategoryName + "-" + strings.Split(c.ID.String(), "-")[0])
			}
		}
	}
	return
}
