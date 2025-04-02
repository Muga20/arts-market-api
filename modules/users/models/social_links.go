package models

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialLink struct to represent a social media link
type SocialLink struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey;default:(UUID())" json:"id"`
	UserID   uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	Platform string    `gorm:"type:varchar(50);not null" json:"platform"`
	Link     string    `gorm:"type:varchar(255);not null" json:"link"`

	// Foreign key relation
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook to generate UUID if not set
func (s *SocialLink) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	// Validate the link URL before saving
	if err := validateURL(s.Link); err != nil {
		return err
	}
	return nil
}

// validateURL checks if the link is safe and well-formed
func validateURL(link string) error {
	// Ensure URL uses HTTP or HTTPS protocol
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return fmt.Errorf("invalid URL protocol, must use 'http' or 'https'")
	}

	// Parse the URL
	parsedURL, err := url.ParseRequestURI(link)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Additional check for harmful domains (for example, block URLs with harmful domains)
	allowedDomains := []string{
		"facebook.com", "instagram.com", "twitter.com", "x.com", "linkedin.com",
		"github.com", "gitlab.com", "bitbucket.org", "youtube.com", "tiktok.com",
		"reddit.com", "discord.com", "telegram.org", "whatsapp.com",
		"medium.com", "dev.to", "stackoverflow.com", "quora.com",
		"pkg.go.dev", "golang.org", "aws.amazon.com", "azure.com", "google.com",
		"apple.com", "microsoft.com", "yahoo.com", "bing.com", "netflix.com",
		"twitch.tv", "pinterest.com", "snapchat.com", "weibo.com", "vk.com",
		"dribbble.com", "behance.net", "unsplash.com", "pexels.com", "flickr.com",
	}

	if !isDomainAllowed(parsedURL.Host, allowedDomains) {
		return fmt.Errorf("URL domain is not allowed")
	}

	return nil
}

// isDomainAllowed checks if the domain of the URL is in the allowed list
func isDomainAllowed(domain string, allowedDomains []string) bool {
	for _, allowedDomain := range allowedDomains {
		if strings.Contains(domain, allowedDomain) {
			return true
		}
	}
	return false
}
