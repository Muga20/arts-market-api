package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/handlers/public"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"

	"gorm.io/gorm"
)

func SetupPublicProfileRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	accountGroup := apiGroup.Group("/profile")

	// Use authentication middleware
	accountGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	// Profile route: Get user profile
	accountGroup.Get("/:identifier", public.GetUserProfileHandler(db, responseHandler))
}
