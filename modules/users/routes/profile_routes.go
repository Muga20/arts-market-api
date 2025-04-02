package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/config" // Importing the config package for Cloudinary
	"github.com/muga20/artsMarket/modules/users/handlers/account"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"github.com/muga20/artsMarket/pkg/middleware"

	"gorm.io/gorm"
)

func SetupAccountRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler, cld *config.CloudinaryClient) {
	accountGroup := apiGroup.Group("/account")

	// Use authentication middleware
	accountGroup.Use(middleware.AuthMiddleware(db, responseHandler))

	// Profile route: Get user profile
	accountGroup.Get("/profile", account.ProfileHandler(db, responseHandler))
	accountGroup.Put("/profile", account.UpdateProfile(db, responseHandler))
	accountGroup.Put("/location", account.UpdateLocation(db, responseHandler))

	// Image upload & removal routes
	accountGroup.Put("/image", account.UpdateUserImage(db, cld, responseHandler))
	accountGroup.Delete("/image", account.RemoveUserImage(db, cld, responseHandler))

	// Privacy settings route
	accountGroup.Get("/privacy-settings", account.GetPrivacySettings(db, responseHandler))
	accountGroup.Put("/privacy-settings", account.UpdatePrivacySettings(db, responseHandler))

	// Create a new social link
	accountGroup.Post("/social-links", account.CreateSocialLink(db, responseHandler))
	accountGroup.Get("/social-links", account.GetSocialLinks(db, responseHandler))
	accountGroup.Put("/social-links/:id", account.UpdateSocialLink(db, responseHandler))
	accountGroup.Delete("/social-links/:id", account.DeleteSocialLink(db, responseHandler))
}
