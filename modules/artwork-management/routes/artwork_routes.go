package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/artworks"
	"github.com/muga20/artsMarket/modules/artwork-management/repository"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// ArtWorksRoutes sets up artwork-related routes
func ArtWorksRoutes(apiGroup fiber.Router, db *gorm.DB, cld *config.CloudinaryClient, responseHandler *handlers.ResponseHandler) {
	artWork := apiGroup.Group("/artworks")

	// Apply the AuthMiddleware to all routes under /artworks
	artWork.Use(middleware.AuthMiddleware(db, responseHandler))

	// Initialize repository
	artworkRepo := repository.NewArtworkRepository(db)

	// Artwork Management Endpoints
	artWork.Post("/", artworks.CreateArtworkHandler(db, cld, responseHandler, artworkRepo))
}
