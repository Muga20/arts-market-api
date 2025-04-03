package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/attributes"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupAttributeRoutes sets up attribute-related routes
func SetupAttributeRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	attributeGroup := apiGroup.Group("/attributes")
	attributeGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	attributeGroup.Post("/", attributes.CreateAttributeHandler(db, responseHandler))
	attributeGroup.Put("/:id", attributes.UpdateAttributeHandler(db, responseHandler))
	attributeGroup.Delete("/:id", attributes.DeleteAttributeHandler(db, responseHandler))
	attributeGroup.Get("/:id", attributes.GetAttributeByIDHandler(db, responseHandler))
	attributeGroup.Get("/", attributes.GetAllAttributesHandler(db, responseHandler))
}
