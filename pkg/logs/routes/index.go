package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
	// Import logs route setup
)

// SetupRoutes initializes all the routes for the app
func LogsModuleSetupRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {

	SetupLogRoutes(apiGroup, db, responseHandler)

}
