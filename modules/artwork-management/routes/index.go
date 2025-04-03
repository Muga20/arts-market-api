package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// SetupRoutes initializes all the routes for the app
func ArtsManagementSetupRoutes(apiGroup fiber.Router, db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) {

	// Authentication-related routes
	SetupTagRoutes(apiGroup, db, responseHandler)
	SetupCategoryRoutes(apiGroup, db, responseHandler)
	SetupAttributeRoutes(apiGroup, db, responseHandler)
	SetupMediumRoutes(apiGroup, db, responseHandler)
	SetupTechniqueRoutes(apiGroup, db, responseHandler)
	SetupCollectionRoutes(apiGroup, db, cld, responseHandler)

}
