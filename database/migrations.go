package database

import (
	"fmt"
	"log"

	// User module imports
	notification "github.com/muga20/artsMarket/modules/notifications/models"
	blocked_user "github.com/muga20/artsMarket/modules/users/models"
	follower "github.com/muga20/artsMarket/modules/users/models"
	roles "github.com/muga20/artsMarket/modules/users/models"
	social_link "github.com/muga20/artsMarket/modules/users/models"
	user_detail "github.com/muga20/artsMarket/modules/users/models"
	user_location "github.com/muga20/artsMarket/modules/users/models"
	user_privacy "github.com/muga20/artsMarket/modules/users/models"
	user_role "github.com/muga20/artsMarket/modules/users/models"
	user_security "github.com/muga20/artsMarket/modules/users/models"
	user_session "github.com/muga20/artsMarket/modules/users/models"
	users "github.com/muga20/artsMarket/modules/users/models"
	error_log "github.com/muga20/artsMarket/pkg/logs/models"

	// Artwork module imports
	artwork_view "github.com/muga20/artsMarket/modules/artwork-management/models/analytics"
	artwork "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	artwork_edition "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	artwork_image "github.com/muga20/artsMarket/modules/artwork-management/models/artWork"
	artwork_attribute "github.com/muga20/artsMarket/modules/artwork-management/models/arttributes"
	artwork_category "github.com/muga20/artsMarket/modules/artwork-management/models/category"
	category "github.com/muga20/artsMarket/modules/artwork-management/models/category"
	collection "github.com/muga20/artsMarket/modules/artwork-management/models/collection"
	artwork_comment "github.com/muga20/artsMarket/modules/artwork-management/models/engagement"
	artwork_favorite "github.com/muga20/artsMarket/modules/artwork-management/models/engagement"
	artwork_like "github.com/muga20/artsMarket/modules/artwork-management/models/engagement"
	medium "github.com/muga20/artsMarket/modules/artwork-management/models/medium"
	artwork_tag "github.com/muga20/artsMarket/modules/artwork-management/models/tags"
	tag "github.com/muga20/artsMarket/modules/artwork-management/models/tags"
	technique "github.com/muga20/artsMarket/modules/artwork-management/models/technique"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	migrations := []interface{}{
		// User module
		&roles.Role{},
		&users.User{},
		&user_session.UserSession{},
		&user_security.UserSecurity{},
		&user_role.UserRole{},
		&user_privacy.UserPrivacySetting{},
		&user_location.UserLocation{},
		&user_detail.UserDetail{},
		&social_link.SocialLink{},
		&follower.Follower{},
		&blocked_user.BlockedUser{},

		// Error logs
		&error_log.ErrorLog{},

		// Notification
		&notification.Notification{},

		// Artwork analytics
		&artwork_view.ArtworkView{},

		// Artwork attributes
		&artwork_attribute.ArtworkAttribute{},

		// Artwork core models
		&artwork.Artwork{},
		&artwork_edition.Edition{},
		&artwork_image.ArtworkImage{},

		// Categories
		&category.Category{},
		&artwork_category.ArtworkCategory{},

		// Collections
		&collection.Collection{},

		// Engagement
		&artwork_comment.ArtworkComment{},
		&artwork_like.ArtworkLike{},
		&artwork_favorite.ArtworkFavorite{},

		// Medium
		&medium.Medium{},

		// Tags
		&tag.Tag{},
		&artwork_tag.ArtworkTag{},

		// Technique
		&technique.Technique{},
	}

	for _, model := range migrations {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("auto migration failed for model %T: %w", model, err)
		}
		log.Printf("✅ Successfully migrated model: %T", model)
	}

	log.Println("✅ Database migration completed successfully")
	return nil
}

func tableExists(db *gorm.DB, tableName string) bool {
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count)
	return count > 0
}
