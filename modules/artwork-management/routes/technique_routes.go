package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/technique"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupTechniqueRoutes sets up technique-related routes
func SetupTechniqueRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	techniqueGroup := apiGroup.Group("/techniques")
	techniqueGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	techniqueGroup.Post("/", technique.CreateTechniqueHandler(db, responseHandler))
	techniqueGroup.Put("/:id", technique.UpdateTechniqueHandler(db, responseHandler))
	techniqueGroup.Delete("/:id", technique.DeleteTechniqueHandler(db, responseHandler))
	techniqueGroup.Get("/:id", technique.GetTechniqueByIDHandler(db, responseHandler))
	techniqueGroup.Get("/", technique.GetAllTechniquesHandler(db, responseHandler))
}
