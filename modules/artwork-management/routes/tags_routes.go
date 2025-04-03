package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/artwork-management/handlers/tags"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"
	"gorm.io/gorm"
)

// SetupTagRoutes sets up tag-related routes
func SetupTagRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	tagGroup := apiGroup.Group("/tags")

	// Apply the AuthMiddleware to all routes under /tags
	tagGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	// Tag Management Endpoints
	tagGroup.Post("/", tags.CreateTagHandler(db, responseHandler))
	tagGroup.Put("/:id", tags.UpdateTagHandler(db, responseHandler))
	tagGroup.Put("/:id/toggle-active", tags.ToggleTagActiveHandler(db, responseHandler))
	tagGroup.Get("/:id", tags.GetTagByIDHandler(db, responseHandler))
	tagGroup.Get("/", tags.GetAllTagsHandler(db, responseHandler))
	tagGroup.Delete("/:id", tags.DeleteTagHandler(db, responseHandler))

}
