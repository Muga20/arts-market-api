package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/medium"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupMediumRoutes sets up medium-related routes
func SetupMediumRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	mediumGroup := apiGroup.Group("/mediums")
	mediumGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	mediumGroup.Post("/", medium.CreateMediumHandler(db, responseHandler))
	mediumGroup.Put("/:id", medium.UpdateMediumHandler(db, responseHandler))
	mediumGroup.Delete("/:id", medium.DeleteMediumHandler(db, responseHandler))
	mediumGroup.Get("/:id", medium.GetMediumByIDHandler(db, responseHandler))
	mediumGroup.Get("/", medium.GetAllMediumsHandler(db, responseHandler))
}
