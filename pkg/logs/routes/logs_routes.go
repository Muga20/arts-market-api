package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// SetupLogRoutes sets up log-related routes
func SetupLogRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	logGroup := apiGroup.Group("/logs")

	// Get logs with pagination
	logGroup.Get("/", handlers.GetLogsHandler(db, responseHandler))

	// Delete logs based on date or range
	logGroup.Delete("/", handlers.DeleteLogsHandler(db, responseHandler))
}
