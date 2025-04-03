package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
	// Import logs route setup
)

// SetupRoutes initializes all the routes for the app
func UserModuleSetupRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler, cld *config.CloudinaryClient) {

	// Authentication-related routes
	SetupAuthRoutes(apiGroup, db, responseHandler)
	SetupAccountRoutes(apiGroup, db, responseHandler, cld)
	SetupRoleRoutes(apiGroup, db, responseHandler)
	SetupAccountSecurityRoutes(apiGroup, db, responseHandler)
	RegisterEngagementRoutes(apiGroup, db, responseHandler)
	SetupPublicProfileRoutes(apiGroup, db, responseHandler)
}
