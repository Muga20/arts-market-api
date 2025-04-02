package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muga20/artsMarket/modules/users/auth"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
	"gorm.io/gorm"
)

// SetupAuthRoutes sets up authentication-related routes
func SetupAuthRoutes(apiGroup fiber.Router, db *gorm.DB, responseHandler *handlers.ResponseHandler) {
	authGroup := apiGroup.Group("/auth")

	// Existing signup route
	authGroup.Post("/signup", auth.SignupHandler(db, responseHandler))
	authGroup.Post("/login", auth.LoginHandler(db, responseHandler))
	authGroup.Post("/logout", auth.LogoutHandler(db, responseHandler))
	authGroup.Post("/logout-other-devices", auth.LogoutOtherDevicesHandler(db, responseHandler))
	authGroup.Post("/reset-password-request", auth.ResetPasswordRequestHandler(db, responseHandler))
	authGroup.Post("/reset-password", auth.ResetPasswordHandler(db, responseHandler))
}
