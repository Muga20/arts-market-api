package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/collection"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupCollectionRoutes sets up collection-related routes
func SetupCollectionRoutes(apiGroup fiber.Router, db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) {
	collectionGroup := apiGroup.Group("/collections")
	collectionGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	// Basic CRUD operations
	collectionGroup.Post("/", collection.CreateCollectionHandler(db, responseHandler))
	collectionGroup.Put("/:id", collection.UpdateCollectionHandler(db, responseHandler))
	collectionGroup.Delete("/:id", collection.DeleteCollectionHandler(db, responseHandler))
	collectionGroup.Get("/:id", collection.GetCollectionByIDHandler(db, responseHandler))
	collectionGroup.Get("/", collection.GetAllCollectionsHandler(db, responseHandler))

	// Status management
	collectionGroup.Put("/:id/status/:status", collection.UpdateCollectionStatusHandler(db, responseHandler))

	// Image uploads
	collectionGroup.Put("/:id/images", collection.UpdateCollectionImagesHandler(db, cld, responseHandler))
	collectionGroup.Delete("/:id/images", collection.RemoveCollectionImageHandler(db, cld, responseHandler))
}
