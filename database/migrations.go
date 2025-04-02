package database

import (
	"fmt"
	"log"

	// Import models from their respective packages
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

	"gorm.io/gorm"
)

// tableExists checks if a table exists in the database
func tableExists(db *gorm.DB, tableName string) bool {
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&count)
	return count > 0
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	// Check if the 'error_logs' table exists to avoid unnecessary migrations
	if tableExists(db, "error_logs") {
		log.Println("✅ Table 'error_logs' already exists, skipping migration")
	} else {
		log.Println("🔄 Creating 'error_logs' table...")
	}

	log.Println("🔄 Running database migrations...")

	// Run migrations for all models and log each migration
	migrations := []interface{}{
		// Module User
		&roles.Role{},                      // Role model
		&users.User{},                      // User model
		&user_session.UserSession{},        // User session model
		&user_security.UserSecurity{},      // User security model
		&user_role.UserRole{},              // User role model
		&user_privacy.UserPrivacySetting{}, // Privacy settings model
		&user_location.UserLocation{},      // Location model
		&user_detail.UserDetail{},          // User details model
		&social_link.SocialLink{},          // Social links model
		&follower.Follower{},               // Followers model
		&blocked_user.BlockedUser{},        // Blocked users model

		// Error logs
		&error_log.ErrorLog{}, // Logging model

		// Add the Notification model
		&notification.Notification{}, // Notification model
	}

	// Loop through all migrations and apply them
	for _, model := range migrations {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("auto migration failed for model %T: %w", model, err)
		}
		log.Printf("✅ Successfully migrated model: %T", model)
	}

	log.Println("✅ Database migration completed successfully")
	return nil
}
